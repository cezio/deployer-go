package deployer

import (
	"io"
	"net/http"
	"strings"
	//"log"
)

func handleDeployment(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	deployment_conf := path[len(path)-1]

	var err = run_config(deployment_conf)
	if err != nil {
		if err.IsMissingConfig() {
			io.WriteString(w, "Config "+deployment_conf+" not found: "+*err.Error())
			return
		}
		if err.IsReadError() {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Config "+deployment_conf+" error: "+*err.Error())
		}
	} else {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "OK")
	}
	return
}

// path to read configs
var directory_path string

func MakeMux(directory string) *http.ServeMux {
	directory_path = directory
	mux := http.NewServeMux()
	mux.HandleFunc("/incoming/", handleDeployment)
	return mux
}
