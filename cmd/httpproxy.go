package cmd

import (
	"log"
	"net/http"
	"strconv"
)

func httpModifyResponse(response *http.Response) error {
	log.Println(*response.Request)
	rpcBalancerUpstreamHttpRequest.WithLabelValues(response.Request.URL.String()[:len(response.Request.URL.String())-1], strconv.Itoa(response.StatusCode)).Inc()
	return nil
}

func httpErrorHandler(rw http.ResponseWriter, req *http.Request, err error) {
	rpcBalancerUpstreamHttpRequest.WithLabelValues(req.URL.String()[:len(req.URL.String())-1], "502").Inc()
	rw.WriteHeader(http.StatusBadGateway)
}