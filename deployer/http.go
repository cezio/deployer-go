package deployer

import (
    "net/http"
    "strings"
    "io"
    //"log"
)




func handleDeployment(w http.ResponseWriter, r *http.Request){
    path := strings.Split(strings.Trim(r.URL.Path, "/"), "/");
    deployment_conf := path[len(path)-1];

    conf, err := NewConfigFromEnv();
    if (err != nil){
        w.WriteHeader(http.StatusNotFound);
        io.WriteString(w, "Config "+deployment_conf + " not found: "+ err.Error());
        return;
    }
    cerr := conf.Read(deployment_conf);
    if (cerr != nil){
        w.WriteHeader(http.StatusInternalServerError);
        io.WriteString(w, "Config "+deployment_conf + " error: " + cerr.Error());
        return;
        }
    conf.Run();
    w.WriteHeader(http.StatusOK);
    io.WriteString(w, "OK");
    return;
}


func MakeMux() (*http.ServeMux){
    mux := http.NewServeMux();
    mux.HandleFunc("/incoming/", handleDeployment);
    return mux;
}
