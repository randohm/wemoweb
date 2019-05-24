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

func Discover(config Config_t) {
    deviceList := make(map[string]map[string]string)
    ctx := context.Background()
    api, _ := wemo.NewByInterface(config.EthDevice)
    devs, _ := api.DiscoverAll(time.Duration(config.DiscoveryTimeout)*time.Second)

    fmt.Println("Discovering devices")
    i := 1
    for _, device := range devs {
        deviceInfo, _ := device.FetchDeviceInfo(ctx)
        fmt.Printf("%2d Found %s %s\n", i, device.Host, deviceInfo.FriendlyName)
        deviceList[deviceInfo.FriendlyName] = map[string]string{"ip_port": device.Host}
        i++
    }

    fmt.Print("Save to file? (y/n)[n] ")
    reader := bufio.NewReader(os.Stdin)
    readIn, _ := reader.ReadString('\n')
    readIn = strings.TrimSuffix(readIn, "\n")
    if readIn == "y" {
        fmt.Printf("Saving to file '%s'\n", config.DevicesFile)
        err := WriteDevices(config, deviceList)
        if err!= nil {
            panic(err)
        }
    }
}
