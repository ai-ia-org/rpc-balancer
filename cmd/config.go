package cmd

import (
	"os"
	"log"
	"gopkg.in/yaml.v3"
)

type Configuration struct {
	Networks []struct {
		ChainId string `yaml:"chainId"`
		Name string `yaml:"name"`
		Path string `yaml:"path"`
		Upstreams []struct {
			Name string `yaml:"name"`
			Url string `yaml:"url"`
			WsUrl string `yaml:"wsUrl"`
		} `yaml:"upstreams"`
	} `yanl:"networks"`
}

func getConfig(configfile *string) Configuration {
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