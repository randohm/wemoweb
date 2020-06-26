package wemoweb

import (
    "io/ioutil"
    "gopkg.in/yaml.v2"
    "log"
)



func ReadDevices(config Config_t) (map[string]map[string]string, error) {
    devicesYaml, err := ioutil.ReadFile(config.DevicesFile)
    if err != nil {
        return map[string]map[string]string{}, err
    }

    var deviceList map[string]map[string]string
    err = yaml.Unmarshal(devicesYaml, &deviceList)
    return deviceList, nil
}



func WriteDevices(config Config_t, deviceList map[string]map[string]string) error {
    yamlOut, err := yaml.Marshal(deviceList)
    if err != nil {
        return err
    }

    err = ioutil.WriteFile(config.DevicesFile, yamlOut, 0644)
    if err != nil {
        return err
    }
    return nil
}



func UpdateDevices(config Config_t, deviceList, newDevices map[string]map[string]string) bool {
    changed := false
    for k, v := range newDevices {
        _, ok := deviceList[k]
        if !ok {
            deviceList[k] = map[string]string{"Host": v["Host"]}
            changed = true
        } else if v["Host"] != deviceList[k]["Host"] {
            log.Printf("Found updated device: %s old:%s new: %s\n", k, deviceList[k], v)
            //deviceList[k]["Host"] = v["Host"]
            deviceList[k] = v
            changed = true
        }
    }
    return changed
}
