package main

import (
	"aproxy/config"
	"flag"
	"log"
	"os"
)

var configPath string

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	flag.StringVar(&configPath, "c", "config.json", "filepath of config.json")
	flag.Parse()
}

func main() {
	c, err := config.ParseConfigFromFile(configPath)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	c.Run()
}
