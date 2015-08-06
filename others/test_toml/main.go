package main

import (
	"log"
	"flag"
	"github.com/BurntSushi/toml"
)

var (
	cfg = flag.String("cfg", "", "Toml file path")
)

func init() {
	flag.Parse()
}

func main() {
	log.Printf("config-path=[%s]", *cfg)
	var config interface{}
	_, err := toml.DecodeFile(*cfg, &config)
	log.Printf("config=[%v], err=[%v]", config, err)
	mysqlDb := config.(map[string]interface{})["db"].(map[string]interface{})["mysql"]
	for k, v := range mysqlDb.(map[string]interface{}) {
		log.Printf("key=[%v], value=[%+v]", k, v)
	}
}
