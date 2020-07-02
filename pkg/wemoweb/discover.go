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

func DiscoverCli(config Config_t) {
    deviceList, err := Discover(config)
    if err != nil {
        panic(err)
    }

    i := 1
    for _, v := range deviceList {
        fmt.Printf("%2d Found %s %s\n", i, v["Host"], v["FriendlyName"])
        //fmt.Printf("    %+v\n", v)
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
        err := WriteDevices(config, deviceList)
        if err!= nil {
            panic(err)
        }
    }
}



func Discover(config Config_t) (map[string]map[string]string, error) {
    deviceList := make(map[string]map[string]string)
    ctx := context.Background()
    api, err := wemo.NewByInterface(config.EthDevice)
    if err != nil {
        return map[string]map[string]string{}, err
    }
    devs, err := api.DiscoverAll(time.Duration(config.DiscoveryTimeout)*time.Second)
    if err != nil {
        return map[string]map[string]string{}, err
    }

    for _, device := range devs {
        deviceInfo, _ := device.FetchDeviceInfo(ctx)
        //fmt.Printf("%+v\n    %+v\n", device, deviceInfo)
        deviceList[deviceInfo.MacAddress] = map[string]string{"Host": device.Host, "FriendlyName": deviceInfo.FriendlyName, "DeviceType": deviceInfo.DeviceType, "SerialNumber": deviceInfo.SerialNumber}
    }

    return deviceList, nil
}
