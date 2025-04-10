package cmd

import (
	"net/http"
	"net/url"
	"strconv"
	"io"
	"log"
	"strings"
	"encoding/json"
)


type rpcEndpoint struct {
	Name string
  Url string
	Remote url.URL
	WsUrl string
	WsRemote url.URL
}

func (r *rpcEndpoint) init() {
	remote, err := url.Parse(r.Url)
	if err != nil {
		log.Println(r.Name, " RPC address is unparsable ",err)
		return
	}
	r.Remote = *remote
	wsRemote, wserr := url.Parse(r.WsUrl)
	if wserr != nil {
		log.Println(r.Name, " WS RPC address is unparsable ",wserr)
		return
	}
	r.WsRemote = *wsRemote
}

func rpcRequestBody(rpc rpcEndpoint, body io.Reader, httpClient http.Client) []byte {
	req, _ := http.NewRequest("POST", rpc.Url, body)
	req.Header.Add("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println("Error connecting RPC node name " + rpc.Name + " with URL " + rpc.Url)
		return nil
	}
	defer resp.Body.Close()
	resp_body, err := io.ReadAll(resp.Body)
	return resp_body
}

func getLatestBlock(rpc rpcEndpoint, httpClient http.Client) string {
	req_body := strings.NewReader(`{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`)
	resp_body := rpcRequestBody(rpc, req_body, httpClient)
	if resp_body == nil {
		log.Println("Unable to get latest block for " + rpc.Url)
		return "0x0"
	}
	var result map[string]any
	err := json.Unmarshal(resp_body, &result)
	if err != nil {
		log.Println("Unable to decode JSON for latest block from " + rpc.Url)
		return "0x0"
	}
	if result["result"] == nil {
		log.Println("Incorrect RPC response for " + rpc.Url)
		return "0x0"
	}
	return result["result"].(string)
}
func getLatestBlockTimestamp(rpc rpcEndpoint, block string, httpClient http.Client) int64 {
	req_body := strings.NewReader(`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["`+block+`", true],"id":1}`)
	resp_body := rpcRequestBody(rpc, req_body, httpClient)
	if resp_body == nil {
		return 0
	}
	var res map[string]interface{}
	err := json.Unmarshal(resp_body, &res)
	if err != nil {
		log.Println("Unable to decode JSON for latest block timestamp from " + rpc.Url)
		return 0
	}
	if res["result"] == nil {
		log.Println("[block-timestamp]: Empty result returned by " + rpc.Url + " and block " + block)
		return 0
	}
	result := res["result"].(map[string]interface{})
	timestamp, _ := strconv.ParseInt(strings.Replace(result["timestamp"].(string), "0x", "", -1), 16, 64)
	return timestamp
}
