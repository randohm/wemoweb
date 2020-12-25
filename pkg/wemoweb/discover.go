package wemoweb

import (
    "fmt"
    "github.com/randohm/go.wemo"
    "context"
    "os"
    "time"
    "bufio"
    "strings"
)



/*
    Discovers running devices and saves results to a file, overwiting the file completely.
*/
func DiscoverCli() {
    deviceList, err := discover()
    if err != nil {
        panic(err)
    }

    i := 1
    for _, v := range deviceList {
        fmt.Printf("%2d Found %s %s\n", i, v["Host"], v["FriendlyName"])
        i++
    }

    fmt.Print("Save to file? (y/n)[n] ")
    reader := bufio.NewReader(os.Stdin)
    readIn, err := reader.ReadString('\n')
    if err != nil {
        panic(err)
    }
    readIn = strings.TrimSuffix(readIn, "\n")
    if readIn == "y" {
        fmt.Printf("Saving to file '%s'\n", config.DevicesFile)
        err := writeDevices(deviceList)
        if err!= nil {
            panic(err)
        }
    }
}



func discover() (map[string]map[string]string, error) {
    deviceList := make(map[string]map[string]string)
    ctx := context.Background()
    api, err := wemo.NewByInterface(config.EthDevice)
    if err != nil {
        log.Errorf("%s", err)
        return map[string]map[string]string{}, err
    }
    devs, err := api.DiscoverAll(time.Duration(config.DiscoveryTimeout)*time.Second)
    if err != nil {
        log.Errorf("%s", err)
        return map[string]map[string]string{}, err
    }

    for _, device := range devs {
        deviceInfo, _ := device.FetchDeviceInfo(ctx)
        deviceList[deviceInfo.MacAddress] = map[string]string{"Host": device.Host, "FriendlyName": deviceInfo.FriendlyName, "DeviceType": deviceInfo.DeviceType, "SerialNumber": deviceInfo.SerialNumber}
    }

    return deviceList, nil
}
