package cmd

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
)

func httpModifyResponse(response *http.Response) error {
	log.Println(*response.Request)
	rpcBalancerUpstreamHttpRequest.WithLabelValues(response.Request.URL.String()[:len(response.Request.URL.String())-1], strconv.Itoa(response.StatusCode)).Inc()
	return nil
}

type proxyFailureFlag struct {
	failed bool
}

type ctxKey int

const proxyFailureCtxKey ctxKey = 0

// httpErrorHandler is invoked by httputil.ReverseProxy when the round trip to
// the upstream fails at the transport level (connection refused/reset, dial
// or read timeout, TLS error, etc.) rather than when the upstream returns a
// legitimate HTTP error status. It records the failure onto the per-attempt
// flag (set up by serveHTTPWithRetry) so the caller can retry against a
// different upstream instead of surfacing this 502 straight to the client.
func httpErrorHandler(rw http.ResponseWriter, req *http.Request, err error) {
	rpcBalancerUpstreamHttpRequest.WithLabelValues(req.URL.String()[:len(req.URL.String())-1], "502").Inc()
	if flag, ok := req.Context().Value(proxyFailureCtxKey).(*proxyFailureFlag); ok {
		flag.failed = true
	}
	rw.WriteHeader(http.StatusBadGateway)
}

// serveHTTPWithRetry proxies an HTTP JSON-RPC request, retrying against other
// healthy upstreams (up to maxUpstreamAttempts) if the chosen upstream fails
// at the transport level. A single random upstream pick with no retry means
// any transient blip on that one node (too brief for the periodic health
// check to catch) surfaces as a 502 to the caller, even though another
// perfectly healthy upstream was available. Falls back to the configured
// fallback upstream, if any, when there are no healthy upstreams at all.
func serveHTTPWithRetry(w http.ResponseWriter, r *http.Request, chainId string, chainName string, ups *upstreams, fallback *upstream) {
	var bodyBytes []byte
	if r.Body != nil {
		bodyBytes, _ = io.ReadAll(r.Body)
		_ = r.Body.Close()
	}

	candidates := ups.getHealthyUpstreamsShuffled()
	if len(candidates) == 0 {
		if fallback == nil {
			log.Println(r.URL.Path, "doesn't have active upstreams and no fallback configured")
			http.Error(w, "no upstreams available", http.StatusServiceUnavailable)
			return
		}
		log.Println(r.URL.Path, "no healthy upstreams, using fallback")
		candidates = []*upstream{fallback}
	}

	maxAttempts := len(candidates)
	if maxAttempts > maxUpstreamAttempts {
		maxAttempts = maxUpstreamAttempts
	}

	var rec *httptest.ResponseRecorder
	for attempt := 0; attempt < maxAttempts; attempt++ {
		u := candidates[attempt]

		req := r.Clone(r.Context())
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		req.ContentLength = int64(len(bodyBytes))
		req.Host = u.RpcEndpoint.Remote.Host
		req.URL.Path = u.RpcEndpoint.Remote.Path

		flag := &proxyFailureFlag{}
		req = req.WithContext(context.WithValue(req.Context(), proxyFailureCtxKey, flag))

		rec = httptest.NewRecorder()
		u.Proxy.ServeHTTP(rec, req)
		rpcBalancerUpstreamHttpRequestTotal.WithLabelValues(chainId, chainName, u.RpcEndpoint.Name, u.RpcEndpoint.Url).Inc()

		if !flag.failed {
			copyRecordedResponse(w, rec)
			return
		}
		log.Println(u.RpcEndpoint.Url, "failed to proxy request for", r.URL.Path, "- retrying with another upstream")
	}

	// Every attempt failed at the transport level; surface the last one's 502.
	copyRecordedResponse(w, rec)
}

func copyRecordedResponse(w http.ResponseWriter, rec *httptest.ResponseRecorder) {
	for key, values := range rec.Header() {
		w.Header()[key] = values
	}
	w.WriteHeader(rec.Code)
	_, _ = w.Write(rec.Body.Bytes())
}
