package deployer

import (
	"io"
	"net/http"
	"strings"
	//"log"
)

func handleDeployment(w http.ResponseWriter, r *http.Request) {
	var path = strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var deploymentConf = path[len(path)-1]

	var err = runConfig(deploymentConf)
	if err != nil {
		if err.IsMissingConfig() {
			io.WriteString(w, "Config "+deploymentConf+" not found: "+*err.Error())
			return
		}
		if err.IsReadError() {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Config "+deploymentConf+" error: "+*err.Error())
		}
	} else {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "OK")
	}
	return
}

// path to read configs
var configurationBase string

// MakeMux creates new ServerMux object
func MakeMux(directory string) *http.ServeMux {
	configurationBase = directory
	mux := http.NewServeMux()
	mux.HandleFunc("/incoming/", handleDeployment)
	return mux
}
