package deployer

import (
	"io"
	"net/http"
	"strings"
)

func readSecret(c *DeploymentConfig, r *http.Request) (*string, *ConfigError) {
	if c.Secret == nil {
		return nil, nil
	}
	if c.SecretHeader != nil {
		var secret = r.Header.Get(*c.SecretHeader)
		return &secret, nil
	} else {
		var bodyReader = r.Body
		if bodyReader == nil {
			bmsg := "Body is empty"
			return nil, &ConfigError{PreconditionsError, &bmsg}
		}
		var secret = make([]byte, MaxSecretSize)
		var read, rerr = bodyReader.Read(secret)
		if rerr != nil {
			/* rmsg := rerr.Error()
			return deploymentConf, &ConfigError{PreconditionsError, &rmsg}
			*/
		}
		var secretStr = string(secret[:read])
		return &secretStr, nil
	}

}

func runConfigFromRequest(r *http.Request) (string, *ConfigError) {
	var path = strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var deploymentConf = path[len(path)-1]
	var err = runConfig(deploymentConf, r)

	return deploymentConf, err
}

func handleDeployment(w http.ResponseWriter, r *http.Request) {
	var deploymentConf, err = runConfigFromRequest(r)
	if err != nil {
		if err.IsMissingConfig() {
			w.WriteHeader(http.StatusNotFound)
			io.WriteString(w, "Config "+deploymentConf+" not found: "+(*err.Error()))
			io.WriteString(w, "\n")
			return
		}
		if err.IsReadError() {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Config "+deploymentConf+" error: "+(*err.Error()))
			io.WriteString(w, "\n")
			return
		}
		if err.IsExecutionError() {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Config "+deploymentConf+" execution error: "+(*err.Error()))
			io.WriteString(w, "\n")
			return
		}
		if err.IsPreconditionsError() {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "Config "+deploymentConf+" preconditions error: "+(*err.Error()))
			io.WriteString(w, "\n")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Config "+deploymentConf+" unknown error: "+(*err.Error()))
		io.WriteString(w, "\n")
		return

	}
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "OK")
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
