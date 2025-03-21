package cmd

import (
	"net/http"
	"net/url"
	"log"
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
	var ethMainnetUpstreams upstreams
	for _, upstream := range config.Upstreams {
		upstreamRpc := rpcEndpoint {Name: upstream.Name, Url: upstream.Url}
		upstreamRpc.init()
		ethMainnetUpstreams.addUpstream(upstreamRpc)
	}
	ethMainnetUpstreams.init()
	handler := func() func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			u := ethMainnetUpstreams.getNextUpstream()
			log.Println(r.URL, u.RpcEndpoint.Url)
			remote, err := url.Parse(u.RpcEndpoint.Url)
			if err != nil {
				panic(err)
			}
			r.Host = remote.Host
			u.Proxy.ServeHTTP(w, r)
		}
	}

	http.HandleFunc("/", handler())
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}