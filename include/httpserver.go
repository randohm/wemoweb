package wemoweb

import (
    "net/http"
    "html/template"
    "log"
    "fmt"
    "crypto/md5"
    "github.com/srandawa/go.wemo"
    "golang.org/x/net/context"
    "io/ioutil"
    "encoding/json"
)



var config_g Config_t
var users map[string]string



func ReadUsers() (map[string]string, error) {
    log.Println("Reading in users file")
    usersJson, err := ioutil.ReadFile(config_g.UsersFile)
    if err != nil {
        return map[string]string{}, err
    }

    var users map[string]string
    err = json.Unmarshal(usersJson, &users)
    return users, nil
}



func HttpLog(r *http.Request) {
    log.Printf("%s %s %s %s (%s)", r.RemoteAddr, r.Method, r.URL, r.Proto, r.UserAgent())
}



func CheckUserPass(user, pass string) bool {
    if users == nil {
        var err error
        users, err = ReadUsers()
        if err != nil {
            return false
        }
    }
    passMd5 := fmt.Sprintf("%x", md5.Sum([]byte(pass)))
    if users[user] == passMd5 {
        return true
    }

    return false
}



func CheckHttpAuth(w http.ResponseWriter, r *http.Request) bool {
    user, pass, ok := r.BasicAuth()
    if !ok || !CheckUserPass(user, pass) {
        w.Header().Set("WWW-Authenticate", `Basic realm="wemoweb"`)
        w.WriteHeader(401)
        w.Write([]byte("Unauthorized.\n"))
        return false
    }
    return true
}



func GenerateRootPage(w http.ResponseWriter, devices map[string]map[string]string, message string) {
    ctx := context.Background()

    for k, v := range devices {
        d := &wemo.Device{Host:v["ip_port"]}
        device_info, _ := d.FetchDeviceInfo(ctx)
        devices[k]["info"] = fmt.Sprintf("%+v", device_info)
        devices[k]["state"] = fmt.Sprintf("%d", d.GetBinaryState())
    }

    tmpl, _ := template.ParseFiles(config_g.HtmlTemplate)
    http_data := struct {
        DeviceData map[string]map[string]string
        Message string
    }{
        DeviceData: devices,
        Message: message,
    }
    tmpl.Execute(w, http_data)
}



func HttpHandler(w http.ResponseWriter, r *http.Request) {
    if config_g.UsersFile != "" {
        if !CheckHttpAuth(w, r) {
            return
        }
    }

    var message string
    devicesList, err := ReadDevices(config_g)
    if err != nil {
        log.Println(err)
        return
    }
    op, _ := r.URL.Query()["op"]
    dev, _ := r.URL.Query()["dev"]
    if len(op) > 0 && len(dev) > 0 {
        device := &wemo.Device{Host:devicesList[dev[0]]["ip_port"]}
        switch op[0] {
            case "on":
                log.Printf("Turning on %s\n", dev[0])
                err = device.On()
                if err != nil {
                    message = fmt.Sprintf("Could not turn on %s: %s", dev[0], err)
                } else {
                    message = fmt.Sprintf("Turned on %s", dev[0])
                }
            case "off":
                log.Printf("Turning off %s\n", dev[0])
                err = device.Off()
                if err != nil {
                    message = fmt.Sprintf("Could not turn off %s: %s", dev[0], err)
                } else {
                    message = fmt.Sprintf("Turned off %s", dev[0])
                }
        }
    }
    GenerateRootPage(w, devicesList, message)
    HttpLog(r)
    //fmt.Printf("%+v\n", r)
}



func IconHandler(w http.ResponseWriter, r *http.Request) {
}



func StartHttp (config Config_t) {
    config_g = config
    log.Printf("Starting webserver on port %d\n", config.HttpPort)
    http.HandleFunc("/", HttpHandler)
    http.HandleFunc("/favicon.ico", IconHandler)
    portStr := fmt.Sprintf(":%d", config.HttpPort)
    if config.UseTls == true {
        http.ListenAndServeTLS(portStr, config.TlsCertFile, config.TlsKeyFile, nil)
    } else {
        http.ListenAndServe(portStr, nil)
    }
}
