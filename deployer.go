package main

import (
    "log"
    "net/http"
    "./deployer"
    )



func main(){

    mux := deployer.MakeMux();
    log.Fatal(http.ListenAndServe(":8081", mux));

}
