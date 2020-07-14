/*
    A Wemoweb, a Wemoweb.
*/
package wemoweb

import (
    "flag"
    "github.com/juju/loggo"
    "errors"
)



const loggerName = "wemoweb"



var config Config
var log loggo.Logger



/*
    Application Main() function.
    Sets up CLI flags.
    Sets logging levels.
    Loads config file.
    Executes based on -mode flag

    Returns:
      error
*/
func Main() error {
    // Setup logger
    log = loggo.GetLogger(loggerName)

    // Setup flags
    configFile := flag.String("config", defaultConfigFile, "Configuration file path")
    mode := flag.String("mode", "server", "Run mode: 'server', 'discover', 'schedule'")
    port := flag.Int("port", 0, "Listen port for HTTP(S) service")
    eth := flag.String("eth", "", "Ethernet device to listen on")
    enforce := flag.Bool("enforce", false, "Enforce the schedule. Requires --mode=schedule")
    debug := flag.Int("debug", 0, "Debug level. 0=none 1=debug 2=trace")
    flag.Parse()

    switch *debug {
        case 0:
            log.SetLogLevel(loggo.INFO)
        case 1:
            log.SetLogLevel(loggo.DEBUG)
        case 2:
            log.SetLogLevel(loggo.TRACE)
    }

    // Load config file
    var err error
    config, err = readConfig(*configFile)
    if err != nil {
        log.Criticalf("%s", err)
        return err
    }

    // Override configs with flags
    if *port > 0 {
        config.HttpPort = *port
    }
    if *eth != "" {
        config.EthDevice = *eth
    }

    // Execute according to mode
    if *mode == "server" {
        StartHttp()
    } else if *mode == "discover" {
        DiscoverCli()
    } else if *mode == "schedule" {
        if *enforce {
            EnforceSchedule()
        } else {
            ShowScheduleCli()
        }
    } else {
        msg := "Unsupported mode"
        log.Errorf(msg)
        return errors.New(msg)
    }

    return nil
}
