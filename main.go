package main

import (
    wemoweb "wemoweb/pkg/wemoweb"
    "log"
)



func main() {
    err := wemoweb.Main()
    if err != nil {
        log.Fatal(err)
    }
}
