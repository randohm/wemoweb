package wemoweb

import (
    "io/ioutil"
    "gopkg.in/yaml.v2"
    "os"
)



const defaultConfigFile = "./config.yml"



/*
    Struct for loading and storing the application configuration.
    This is meant to load from a YAML file with corresponding keys.
*/
type Config struct {
    HttpPort int            // TCP port for the HTTP listener to bind
    EthDevice string        // Ethernet device to discover devices on
    DevicesFile string      // Path to YAML file storing device information
    DiscoveryTimeout int    // Timeout on discovery attempts
    HtmlTemplate string     // Path to HTML template file
    UsersFile string        // Path to file containing authentication information
    UseTls bool             // Flag on whether to use plaintext or TLS (HTTP or HTTPS)
    TlsCertFile string      // Path to TLS cert file
    TlsKeyFile string       // Path to TLS key file
    ScheduleFile string     // Path to YAML file containing schedule information
}



func readConfig(configFile string) (Config, error) {
    _, err := os.Stat(configFile)
    if err != nil {
        log.Errorf("%s", err)
        return Config{}, err
    }
    configData, err := ioutil.ReadFile(configFile)
    if err != nil {
        log.Errorf("%s", err)
        return Config{}, err
    }

    var config Config
    err = yaml.Unmarshal(configData, &config)
    log.Tracef("Config: %+v", config)
    return config, nil
}
