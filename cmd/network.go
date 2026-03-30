package cmd

type network struct {
	ChainId  string
	Name     string
	Proxies  *upstreams
	Fallback *upstream
}
