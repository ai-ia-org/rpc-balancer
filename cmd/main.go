package cmd

import (
	"net/http"
	"log"
	"flag"
)

var connectTimeout = 5
var upstreamCheckInterval = 30
var blockHealthyDiff int64 = 5
var timestampHealthyDiff int64 = 3
var config Configuration

func ReverseProxyErrHandle(res http.ResponseWriter, req *http.Request, err error) {
	log.Println(res.Header())
	log.Println(err)
}

func ReverseProxyModifyResponse(res *http.Response) error {
	log.Println(res.StatusCode)
	return nil
}

func Run() {
	configFilename := flag.String("config", "config.yaml", "Configuration file location")
	config = getConfig(configFilename)
	nets :=  make(map[string]network)
	handler := func() func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			net := nets[r.URL.Path]
			upgrade := false
			for _, header := range r.Header["Upgrade"] {
				if header == "websocket" {
					upgrade = true
					break
				}
			}
			if !upgrade {
				u := net.Proxies.getNextUpstream()
				if u == nil {
					log.Println(r.URL.Path, "doesn't have active upstreams")
					return
				}
				r.Host = u.RpcEndpoint.Remote.Host
				r.URL.Path = u.RpcEndpoint.Remote.Path
				u.Proxy.ModifyResponse = ReverseProxyModifyResponse
				u.Proxy.ErrorHandler = ReverseProxyErrHandle
				u.Proxy.ServeHTTP(w, r)
			}	else {
				u := net.Proxies.getNextWsUpstream()
				if u == nil {
					log.Println(r.URL.Path, " doesn't have active upstreams")
					return
				}
				r.Host = u.RpcEndpoint.WsRemote.Host
				r.URL.Path = u.RpcEndpoint.WsRemote.Path
				u.WsProxy.ServeHTTP(w, r)
			}
		}
	}
	for _, net := range config.Networks {
		up := &upstreams{}
		nets[net.Path] = network {ChainId: net.ChainId, Name: net.Name, Proxies: up}
		for _, upstream := range net.Upstreams {
			upstreamRpc := rpcEndpoint {Name: upstream.Name, Url: upstream.Url, WsUrl: upstream.WsUrl}
			upstreamRpc.init()
			up.addUpstream(upstreamRpc)
		}
		up.init()
		http.HandleFunc(net.Path, handler())
	}
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`{"status": "ok"}`))
		if err != nil {
			panic(err)
		}
	})
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}