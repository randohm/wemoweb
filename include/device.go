package wemoweb

import (
    "io/ioutil"
    "encoding/json"
    "log"
)



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



func UpdateDevices(config Config_t, deviceList, newDevices map[string]map[string]string) bool {
    changed := false
    for k, v := range newDevices {
        if v["ip_port"] != deviceList[k]["ip_port"] {
            log.Printf("Found updated device: %s old:%s new: %s\n", k, deviceList[k], v)
            deviceList[k]["ip_port"] = v["ip_port"]
            changed = true
        }
    }
    return changed
}
