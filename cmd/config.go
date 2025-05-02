package cmd

import (
	"flag"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type Configuration struct {
	Networks []struct {
		ChainId   string `yaml:"chainId"`
		Name      string `yaml:"name"`
		Path      string `yaml:"path"`
		Upstreams []struct {
			Name  string `yaml:"name"`
			Url   string `yaml:"url"`
			WsUrl string `yaml:"wsUrl"`
		} `yaml:"upstreams"`
	} `yaml:"networks"`
}

var connectTimeout = 5
var upstreamCheckInterval = 15
var blockHealthyDiff int64 = 5
var timestampHealthyDiff int64 = 3
var metricsPort *int
var port *int
var config Configuration

func getConfig() Configuration {
	configfile := flag.String("config", "config.yaml", "Configuration file location")
	port = flag.Int("port", 8080, "Application port")
	metricsPort = flag.Int("metrics-port", 6060, "Prometheus metrics port")

	flag.Parse()
	config := Configuration{}
	yamlFile, err := os.ReadFile(*configfile)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return config
}
