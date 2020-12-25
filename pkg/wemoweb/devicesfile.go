package wemoweb

import (
    "io/ioutil"
    "gopkg.in/yaml.v2"
)



func readDevices() (map[string]map[string]string, error) {
    devicesYaml, err := ioutil.ReadFile(config.DevicesFile)
    if err != nil {
        log.Errorf("%s", err)
        return nil, err
    }

    var deviceList map[string]map[string]map[string]string
    err = yaml.Unmarshal(devicesYaml, &deviceList)
    log.Tracef("Devices read from file: %+v", deviceList)
    return deviceList["devices"], nil
}



func writeDevices(deviceList map[string]map[string]string) error {
    top := map[string]map[string]map[string]string{"devices": deviceList}
    yamlOut, err := yaml.Marshal(top)
    if err != nil {
        log.Errorf("%s", err)
        return err
    }

    err = ioutil.WriteFile(config.DevicesFile, append([]byte("---\n"), yamlOut...), 0644)
    if err != nil {
        log.Errorf("%s", err)
        return err
    }
    return nil
}



func updateDevices(deviceList, newDevices map[string]map[string]string) bool {
    changed := false
    for k, v := range newDevices {
        _, ok := deviceList[k]
        if !ok {
            deviceList[k] = v
            changed = true
        } else if v["Host"] != deviceList[k]["Host"] || v["FriendlyName"] != deviceList[k]["FriendlyName"] {
            log.Debugf("Found updated device: %s\n", k)
            deviceList[k] = v
            changed = true
        }
    }
    return changed
}
