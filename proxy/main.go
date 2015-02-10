package main

import (
	"flag"
	"log"
	"os"
	"proxy"
)

var (
	flagConfig = flag.String("f", "config.json", "config file for proxy")
)

func main() {
	flag.Parse()
	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}
	proxy := proxy.NewProxy(*flagConfig)
	log.Fatal(proxy.Start())
}
