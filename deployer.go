package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/cezio/deployer-go/deployer"
)

type cliFlags struct {
	Port      int
	Directory string
}

func main() {

	// default flags
	const (
		defaultPort      = 8081
		defaultConfigDir = "."
	)

	// cli parser

	var port = flag.Int("port", defaultPort, "Port to listen on (default: 8081)")
	var configdir = flag.String("configdir", defaultConfigDir, "Directory to read from (default: .)")

	flag.Parse()
	log.Printf("starting with configuration:\n port: %v\n config dir: %v\n", *port, *configdir)

	mux := deployer.MakeMux(*configdir)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), mux))
}
