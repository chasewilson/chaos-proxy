package main

import (
	"flag"
)

func main() {
	configFile := flag.String("config", "", "path to config file")
	flag.Parse()

}
