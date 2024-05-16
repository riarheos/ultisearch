package main

import "flag"

func main() {
	configFile := flag.String("config", "config.yaml", "Path to config file")
	flag.Parse()

	config, err := ReadConfig(*configFile)
	if err != nil {
		panic(err)
	}

	server := NewServer(config)
	err = server.Start()
	if err != nil {
		panic(err)
	}
}
