package main

import (
	"flag"
	"github.com/treant5612/aproxy/config"
	"log"
)

var configPath, key string
var genClientConf, genServerConf bool

func init() {
	log.SetFlags(log.LstdFlags)
	flag.StringVar(&configPath, "conf", "config.txt", "run with specified config \n default ./config.txt")
	flag.BoolVar(&genClientConf, "c", false, "create a client config template")
	flag.BoolVar(&genServerConf, "s", false, "create a server config template")
	flag.StringVar(&key, "key", "defaultKey", "specify a key while creating config.txt")
	flag.Parse()
}

func main() {
	if genConfig() {
		return
	}
	c := config.OpenConfig(configPath)
	if c == nil {
		return
	}
	c.Run()
}

func genConfig() bool {
	if !genServerConf && !genClientConf {
		return false
	}
	confType := 1
	if genClientConf {
		confType = 0
	}
	config.GenerateConfig(confType, key)
	return true
}
