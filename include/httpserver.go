package wemoweb

import (
    "net/http"
    "html/template"
    "log"
    "fmt"
    //"crypto/md5"
    "github.com/savaki/go.wemo"
    "golang.org/x/net/context"
)



var config_g Config_t



func HttpLog(r *http.Request) {
    log.Printf("%s %s %s %s (%s)", r.RemoteAddr, r.Method, r.URL, r.Proto, r.UserAgent())
}



func checkUserPass(user, pass string) bool {
    //passMd5 := fmt.Sprintf("%x", md5.Sum([]byte(pass)))
    return false
}



func checkHttpAuth(w http.ResponseWriter, r *http.Request) bool {
    user, pass, ok := r.BasicAuth()
    if !ok || !checkUserPass(user, pass) {
        w.Header().Set("WWW-Authenticate", `Basic realm="wemoweb"`)
        w.WriteHeader(401)
        w.Write([]byte("Unauthorised.\n"))
        return false
    }
    return true
}



func GenerateRootPage(w http.ResponseWriter, devices map[string]map[string]string) {
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
        ActionResult string
    }{
        DeviceData: devices,
    }
    tmpl.Execute(w, http_data)
}



func HttpHandler(w http.ResponseWriter, r *http.Request) {
    if config_g.UsersFile != "" {
        if !checkHttpAuth(w, r) {
            return
        }
    }

    devicesList, err := ReadDevices(config_g)
    if err != nil {
        log.Println(err)
        return
    }
    op, _ := r.URL.Query()["op"]
    dev, _ := r.URL.Query()["dev"]
    if len(op) > 0 && len(dev) > 0 {
        switch op[0] {
            case "on":
                log.Printf("Turning on %s\n", dev[0])
                device := &wemo.Device{Host:devicesList[dev[0]]["ip_port"]}
                device.On()
            case "off":
                log.Printf("Turning off %s\n", dev[0])
                device := &wemo.Device{Host:devicesList[dev[0]]["ip_port"]}
                device.Off()
        }
    }
    GenerateRootPage(w, devicesList)
    //log.Println(fmt.Sprintf("%+v", r))
    HttpLog(r)
}



func StartHttp (config Config_t) {
    config_g = config
    log.Println("Starting webserver")
    http.HandleFunc("/", HttpHandler)
    portStr := fmt.Sprintf(":%d", config.HttpPort)
    if config.UseTls == true {
        http.ListenAndServeTLS(portStr, config.TlsCertFile, config.TlsKeyFile, nil)
    } else {
        http.ListenAndServe(portStr, nil)
    }
}
