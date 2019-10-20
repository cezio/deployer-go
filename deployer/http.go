package deployer

import (
	"io"
	"net/http"
	"strings"
	//"log"
)

func runConfigFromRequest(r *http.Request) (string, *ConfigError) {
	var path = strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var deploymentConf = path[len(path)-1]
	var bodyReader = r.Body
	if bodyReader == nil {
		bmsg := "Body is empty"
		return deploymentConf, &ConfigError{PreconditionsError, &bmsg}
	}
	var secret = make([]byte, MaxSecretSize)
	var read, rerr = bodyReader.Read(secret)
	if rerr != nil {
		/* rmsg := rerr.Error()
		return deploymentConf, &ConfigError{PreconditionsError, &rmsg}
		*/
	}
	secret = secret[:read]
	var err = runConfig(deploymentConf, r.Method, string(secret))

	return deploymentConf, err
}

func handleDeployment(w http.ResponseWriter, r *http.Request) {
	var deploymentConf, err = runConfigFromRequest(r)
	if err != nil {
		if err.IsMissingConfig() {
			w.WriteHeader(http.StatusNotFound)
			io.WriteString(w, "Config "+deploymentConf+" not found: "+*err.Error())
			return
		}
		if err.IsReadError() {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Config "+deploymentConf+" error: "+*err.Error())
			return
		}
		if err.IsExecutionError() {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Config "+deploymentConf+" execution error: "+*err.Error())
			return
		}
		if err.IsPreconditionsError() {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "Config "+deploymentConf+" preconditions error: "+*err.Error())
			return
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
