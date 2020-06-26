package wemoweb

import (
    "io/ioutil"
    "encoding/json"
)

const (
    defaultConfigFile = "./config.json"
)



type Config_t struct {
    HttpPort int `json:"http_port"`
    EthDevice string `json:"eth_device"`
    DevicesFile string `json:"devices_file"`
    DiscoveryTimeout int `json:"discovery_timeout"`
    HtmlTemplate string `json:"html_tmpl"`
    UsersFile string `json:"users_file"`
    UseTls bool `json:"use_tls"`
    TlsCertFile string `json:"tls_cert_file"`
    TlsKeyFile string `json:"tls_key_file"`
}



func ReadConfig(configFile string) (Config_t, error){
    configJson, err := ioutil.ReadFile(configFile)
    if err != nil {
        return Config_t{}, err
    }

    var config Config_t
    err = json.Unmarshal(configJson, &config)
    return config, nil
}
