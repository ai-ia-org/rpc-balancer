package cmd

import (
	"fmt"
	"log"
	"net/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Run() {
	config = getConfig()
	nets := make(map[string]network)
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
				u.Proxy.ServeHTTP(w, r)
			} else {
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
		nets[net.Path] = network{ChainId: net.ChainId, Name: net.Name, Proxies: up}
		for _, upstream := range net.Upstreams {
			upstreamRpc := rpcEndpoint{Name: upstream.Name, Url: upstream.Url, WsUrl: upstream.WsUrl}
			upstreamRpc.init()
			up.addUpstream(upstreamRpc)
		}
		up.init(net.ChainId, net.Name)
		http.HandleFunc(net.Path, handler())
	}
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`{"status": "ok"}`))
		if err != nil {
			panic(err)
		}
	})
	go func() {
		metricsServer := &http.Server{
			Addr:    fmt.Sprintf(":%d",*metricsPort),
			Handler: promhttp.Handler(),
		}
		log.Println("Starting Prometheus metrics server on", *metricsPort, "port")
		if err := metricsServer.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start metrics server: %v", err)
		}
	}()
	appPortStr := fmt.Sprintf(":%d",*port)
	log.Println("Starting application server on", *port, "port")
	err := http.ListenAndServe(appPortStr, nil)
	if err != nil {
		panic(err)
	}
}
