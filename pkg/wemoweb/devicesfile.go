package wemoweb

import (
    "io/ioutil"
    "gopkg.in/yaml.v2"
    "log"
)



func ReadDevices(config Config_t) (map[string]map[string]string, error) {
    devicesYaml, err := ioutil.ReadFile(config.DevicesFile)
    if err != nil {
        return nil, err
    }

    var deviceList map[string]map[string]map[string]string
    err = yaml.Unmarshal(devicesYaml, &deviceList)
    return deviceList["devices"], nil
}



func WriteDevices(config Config_t, deviceList map[string]map[string]string) error {
    top := map[string]map[string]map[string]string{"devices": deviceList}
    yamlOut, err := yaml.Marshal(top)
    if err != nil {
        return err
    }

    err = ioutil.WriteFile(config.DevicesFile, append([]byte("---\n"), yamlOut...), 0644)
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
            deviceList[k] = v
            changed = true
        } else if v["Host"] != deviceList[k]["Host"] || v["FriendlyName"] != deviceList[k]["FriendlyName"] {
            log.Printf("Found updated device: %s\n", k)
            deviceList[k] = v
            changed = true
        }
    }
    return changed
}
