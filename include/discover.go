package wemoweb

import (
    "fmt"
    "github.com/srandawa/go.wemo"
    "golang.org/x/net/context"
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
    for k, v := range deviceList {
        fmt.Printf("%2d Found %s %s\n", i, v["ip_port"], k)
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
        deviceList[deviceInfo.FriendlyName] = map[string]string{"ip_port": device.Host}
    }

    return deviceList, nil
}
