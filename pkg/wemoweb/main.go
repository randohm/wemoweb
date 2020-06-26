package wemoweb

import (
    "flag"
    "log"
)



func Main() error {
    // Setup flags
    configFile := flag.String("config", defaultConfigFile, "Configuration file path")
    mode := flag.String("mode", "server", "Run mode: 'server', 'discover'")
    port := flag.Int("port", 0, "Listen port for HTTP(S) service")
    eth := flag.String("eth", "", "Ethernet device to listen on")

    flag.Parse()

    // Load config file
    config, err := ReadConfig(*configFile)
    if err != nil {
        log.Fatal(err)
    }
    //log.Printf("%+v", config)

    // Override configs with flags
    if *port > 0 {
        config.HttpPort = *port
    }
    if *eth != "" {
        config.EthDevice = *eth
    }

    // Execute according to mode
    if *mode == "server" {
        StartHttp(config)
    } else if  *mode == "discover" {
        DiscoverCli(config)
    } else {
        log.Fatal("Unsupported mode")
    }

    return nil
}
