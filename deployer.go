package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/cezio/deployer-go/deployer"
)

type cliFlags struct {
	port      int
	directory string
}

func main() {

	var flags = cliFlags{8081, "."}
	flag.Int("port", flags.port, "Port to listen on (default: 8081)")
	flag.String("configdir", flags.directory, "Directory to read from (default: .)")

	flag.Parse()
	mux := deployer.MakeMux(flags.directory)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(flags.port), mux))
}
