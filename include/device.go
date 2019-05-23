package wemoweb

import (
    "io/ioutil"
    "encoding/json"
    //"fmt"
)



type WemoDevice struct {
    Name string `json:"name"`
    IP string `json:"ip"`
    Port int `json:"port"`
}



func ReadDevices(config Config_t) (map[string]map[string]string, error) {
    devicesJson, err := ioutil.ReadFile(config.DevicesFile)
    if err != nil {
        return map[string]map[string]string{}, err
    }

    var deviceList map[string]map[string]string
    err = json.Unmarshal(devicesJson, &deviceList)
    return deviceList, nil
}



func WriteDevices(config Config_t, deviceList map[string]map[string]string) error {
    jsonOut, err := json.Marshal(deviceList)
    if err != nil {
        return err
    }

    err = ioutil.WriteFile(config.DevicesFile, jsonOut, 0644)
    if err != nil {
        return err
    }
    return nil
}
