package wemoweb

import (
    "io/ioutil"
    "gopkg.in/yaml.v2"
)



const defaultConfigFile = "./config.yml"



type Config_t struct {
    HttpPort int `yaml:"HttpPort"`
    EthDevice string `yaml:"EthDevice"`
    DevicesFile string `yaml:"DevicesFile"`
    DiscoveryTimeout int `yaml:"DiscoveryTimeout"`
    HtmlTemplate string `yaml:"HtmlTemplate"`
    UsersFile string `yaml:"UsersFile"`
    UseTls bool `yaml:"UseTls"`
    TlsCertFile string `yaml:"TlsCertFile"`
    TlsKeyFile string `yaml:"TlsKeyFile"`
    ScheduleFile string `yaml:"ScheduleFile"`
}



func ReadConfig(configFile string) (Config_t, error) {
    configData, err := ioutil.ReadFile(configFile)
    if err != nil {
        return Config_t{}, err
    }

    var config Config_t
    err = yaml.Unmarshal(configData, &config)
    return config, nil
}
