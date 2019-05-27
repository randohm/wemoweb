package main

import (
     "./include"
     "log"
     "os"
)



func main() {
    config, err := wemoweb.ReadConfig()
    if err != nil {
        log.Fatal(err)
    }

    if len(os.Args) > 1 {
        if os.Args[1] == "discover" {
            wemoweb.DiscoverCli(config)
        } else if os.Args[1] == "server" {
            wemoweb.StartHttp(config)
        } else {
            log.Fatalf("Unknown command: %s\n", os.Args[1])
        }
    }
}
