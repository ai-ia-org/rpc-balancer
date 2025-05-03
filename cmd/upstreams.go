package cmd

import (
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

type upstream struct {
	Proxy       *httputil.ReverseProxy
	WsProxy     *WebsocketProxy
	RpcEndpoint rpcEndpoint
}

type upstreams struct {
	Upstreams        []upstream
	WsUpstreams      []*upstream
	HealthyUpstreams []*upstream
	HttpClient       http.Client
}

var randomSource *rand.Rand

func (u *upstreams) init(chainId string, chainName string) {
	u.HttpClient = http.Client{
		Timeout: time.Duration(connectTimeout) * time.Second,
	}
	go u.setHealthyUpstreams(chainId, chainName)
	randomSource = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func (u *upstreams) addUpstream(rpc rpcEndpoint) {
	remote, err := url.Parse(rpc.Url)
	if err != nil {
		log.Println(rpc.Name, " RPC address is unparsable ", err)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.ModifyResponse = httpModifyResponse
	proxy.ErrorHandler = httpErrorHandler
	var wsProxy *WebsocketProxy
	if rpc.WsUrl != "" {
		remote, err = url.Parse(rpc.WsUrl)
		if err != nil {
			log.Println(rpc.Name, " WS RPC address is unparsable ", err)
			wsProxy = nil
		} else {
			wsProxy = NewWsProxy(remote)
		}
	}
	up := upstream{Proxy: proxy, WsProxy: wsProxy, RpcEndpoint: rpc}
	u.Upstreams = append(u.Upstreams, up)
}

func (u *upstreams) setHealthyUpstreams(chainId string, chainName string) {
	for {
		upstreamNum := len(u.Upstreams)
		var wg sync.WaitGroup
		blocks := make([]int64, upstreamNum)
		timestamps := make([]int64, upstreamNum)
		wg.Add(upstreamNum)
		for i := 0; i < upstreamNum; i++ {
			go func(i int) {
				blockString := getLatestBlock(u.Upstreams[i].RpcEndpoint, u.HttpClient)
				blocks[i], _ = strconv.ParseInt(strings.ReplaceAll(blockString, "0x", ""), 16, 64)
				timestamps[i] = getLatestBlockTimestamp(u.Upstreams[i].RpcEndpoint, blockString, u.HttpClient)
				rpcBalancerUpstreamLatestBlock.WithLabelValues(chainId, chainName, u.Upstreams[i].RpcEndpoint.Name, u.Upstreams[i].RpcEndpoint.Url).Set(float64(blocks[i]))
				rpcBalancerUpstreamLatestBlockTimestamp.WithLabelValues(chainId, chainName, u.Upstreams[i].RpcEndpoint.Name, u.Upstreams[i].RpcEndpoint.Url).Set(float64(timestamps[i]))
				defer wg.Done()
			}(i)
		}
		wg.Wait()
		var maxBlock, maxTimestamp int64 = 0, 0
		for i := 0; i < upstreamNum; i++ {
			if blocks[i] > maxBlock {
				maxBlock = blocks[i]
				maxTimestamp = timestamps[i]
			}
		}
		rpcBalancerChainLatestBlock.WithLabelValues(chainId, chainName).Set(float64(maxBlock))
		rpcBalancerChainLatestBlockTimestamp.WithLabelValues(chainId, chainName).Set(float64(maxTimestamp))
		for i := 0; i < upstreamNum; i++ {
			if (maxBlock-blocks[i] < blockHealthyDiff || maxTimestamp-timestamps[i] < timestampHealthyDiff) && maxBlock > 0 {
				if !slices.Contains(u.HealthyUpstreams, &u.Upstreams[i]) {
					u.HealthyUpstreams = append(u.HealthyUpstreams, &u.Upstreams[i])
					if u.Upstreams[i].WsProxy != nil {
						u.WsUpstreams = append(u.WsUpstreams, &u.Upstreams[i])
					}
				}
				rpcBalancerUpstreamUp.WithLabelValues(chainId, chainName, u.Upstreams[i].RpcEndpoint.Name, u.Upstreams[i].RpcEndpoint.Url).Set(1)
				log.Println(u.Upstreams[i].RpcEndpoint.Url, "is healthy")
			} else {
				index := slices.Index(u.HealthyUpstreams, &u.Upstreams[i])
				if index != -1 {
					u.HealthyUpstreams = slices.Delete(u.HealthyUpstreams, index, index+1)
				}
				index = slices.Index(u.WsUpstreams, &u.Upstreams[i])
				if index != -1 {
					u.WsUpstreams = slices.Delete(u.WsUpstreams, index, index+1)
				}
				rpcBalancerUpstreamUp.WithLabelValues(chainId, chainName, u.Upstreams[i].RpcEndpoint.Name, u.Upstreams[i].RpcEndpoint.Url).Set(0)
				log.Println(u.Upstreams[i].RpcEndpoint.Url, "is not healthy")
			}
		}
		rpcBalancerChainHealthyUpstreamNum.WithLabelValues(chainId, chainName).Set(float64(len(u.HealthyUpstreams)))
		time.Sleep(time.Duration(upstreamCheckInterval) * time.Second)
	}
}

func (u *upstreams) getNextUpstream() *upstream {
	if len(u.HealthyUpstreams) > 0 {
		n := randomSource.Int() % len(u.HealthyUpstreams)
		return u.HealthyUpstreams[n]
	} else {
		return nil
	}
}

func (u *upstreams) getNextWsUpstream() *upstream {
	if len(u.WsUpstreams) > 0 {
		n := randomSource.Int() % len(u.WsUpstreams)
		return u.WsUpstreams[n]
	} else {
		return nil
	}
}
