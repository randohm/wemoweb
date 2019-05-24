package main

import (
     "./include"
     "log"
     //"fmt"
     "os"
)



func main() {
    config, err := wemoweb.ReadConfig()
    if err != nil {
        log.Fatal(err)
    }
    //fmt.Printf("%+v\n", config)

    if len(os.Args) > 1 {
        if os.Args[1] == "discover" {
            wemoweb.Discover(config)
        } else if os.Args[1] == "server" {
            wemoweb.StartHttp(config)
        } else {
            log.Fatalf("Unknown command: %s\n", os.Args[1])
        }
    }
}
