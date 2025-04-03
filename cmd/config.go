package cmd

import (
	"io/ioutil"
	"log"
	"gopkg.in/yaml.v3"
)

type Configuration struct {
	Upstreams []struct {
		Name string `yaml:"name"`
  	Url string `yaml:"url"`
		WsUrl string `yaml:"wsUrl"`
	} `yaml:"upstreams"`
}

func getConfig(configfile *string) Configuration {
	config := Configuration{}
	yamlFile, err := ioutil.ReadFile(*configfile)
	if err != nil {
			log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return config
}