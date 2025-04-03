package cmd

import (
	"net/http"
	"net/url"
	"flag"
)

var connectTimeout = 5
var upstreamCheckInterval = 30
var blockHealthyDiff int64 = 5
var timestampHealthyDiff int64 = 3
var config Configuration

func Run() {
	configFilename := flag.String("config", "config.yaml", "Configuration file location")
	config = getConfig(configFilename)
	var ethUpstreams upstreams
	for _, upstream := range config.Upstreams {
		upstreamRpc := rpcEndpoint {Name: upstream.Name, Url: upstream.Url, WsUrl: upstream.WsUrl}
		upstreamRpc.init()
		ethUpstreams.addUpstream(upstreamRpc)
	}
	ethUpstreams.init()
	handler := func() func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			upgrade := false
			for _, header := range r.Header["Upgrade"] {
				if header == "websocket" {
					upgrade = true
					break
				}
			}
			if upgrade == false {
				u := ethUpstreams.getNextUpstream()
				remote, err := url.Parse(u.RpcEndpoint.Url)
				if err != nil {
					panic(err)
				}
				r.Host = remote.Host
				u.Proxy.ServeHTTP(w, r)
			}	else {
				u := ethUpstreams.getNextWsUpstream()
				remote, err := url.Parse(u.RpcEndpoint.WsUrl)
				if err != nil {
					panic(err)
				}
				r.Host = remote.Host
				u.WsProxy.ServeHTTP(w, r)
			}
		}
	}
	http.HandleFunc("/", handler())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status": "ok"}`))
	})
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}