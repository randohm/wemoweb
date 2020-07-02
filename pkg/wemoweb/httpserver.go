package wemoweb

import (
    "net/http"
    "html/template"
    "log"
    "fmt"
    "crypto/md5"
    "github.com/randohm/go.wemo"
    "io/ioutil"
    "encoding/json"
    "strconv"
    "time"
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



func HttpLog(r *http.Request, status int) {
    user, _, _ := r.BasicAuth()
    if user == "" {
        user = "-"
    }
    log.Printf("%s %s %d %s %s %s (%s)", r.RemoteAddr, user, status, r.Method, r.URL, r.Proto, r.UserAgent())
    //log.Printf("%+v\n", r)
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
        http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
        HttpLog(r, 401)
        return false
    }
    return true
}



func GenerateRootPage(w http.ResponseWriter, devices map[string]map[string]string, message string) error {
    for _, v := range devices {
        d := &wemo.Device{Host:v["Host"]}
        v["state"] = fmt.Sprintf("%d", d.GetBinaryState())
    }

    tmpl, _ := template.ParseFiles(config_g.HtmlTemplate) // TODO: check for file first
    httpData := struct {
        Mode string
        DeviceData map[string]map[string]string
        NewDeviceData map[string]map[string]string
        Message string
    }{
        Mode: "main",
        DeviceData: devices,
        Message: message,
    }
    err := tmpl.Execute(w, httpData)
    if err != nil {
        log.Printf("Error: %s\n", err)
        return err
    }

    return nil
}



func guiHandler(w http.ResponseWriter, r *http.Request) {
    if config_g.UsersFile != "" {
        if !CheckHttpAuth(w, r) {
            return
        }
    }

    var message string
    devices, err := ReadDevices(config_g)
    if err != nil {
        log.Println(err)
        return
    }
    op, _ := r.URL.Query()["op"]
    dev, _ := r.URL.Query()["dev"]
    length, _ := r.URL.Query()["len"]
    if len(op) > 0 && len(dev) > 0 {
        device := &wemo.Device{Host:devices[dev[0]]["Host"]}
        switch op[0] {
            case "on":
                log.Printf("Turning on %s\n", devices[dev[0]]["FriendlyName"])
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
            case "timer":
                if len(length) > 0 {
                    log.Printf("Turning on %s for %s minutes\n", dev[0], length[0])
                    minutes, _ := strconv.Atoi(length[0])
                    err = device.On()
                    if err != nil {
                        message = fmt.Sprintf("Could not turn on %s: %s", dev[0], err)
                    } else {
                        message = fmt.Sprintf("Turned on %s", dev[0])
                    }
                    go timerOff(device, minutes)
                    message = fmt.Sprintf("Set timer on %s for %s minutes", dev[0], length[0])
                } else {
                    http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
                    HttpLog(r, 500)
                    log.Println("Length not specified")
                    return
                }
        }
    }
    err = GenerateRootPage(w, devices, message)
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        HttpLog(r, 500)
        log.Printf("Error: %s\n", err)
        return
    }

    HttpLog(r, 200)
}



func timerOff(device *wemo.Device, minutes int) {
    sleepTime := time.Duration(minutes) * time.Minute
    log.Printf("Setting timer for %+v\n", sleepTime)
    time.Sleep(sleepTime)
    err := device.Off()
    if err != nil {
        log.Printf("Error turning off after sleep: %s\n", err)
        return
    }
    log.Printf("Turned off after %+v", sleepTime)
}



func iconHandler(w http.ResponseWriter, r *http.Request) {
}



func discoverHandler(w http.ResponseWriter, r *http.Request) {
    message := "No detected device changes"
    devices, err := ReadDevices(config_g)
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        HttpLog(r, 500)
        log.Printf("Error: %s\n", err)
        return
    }
    newDevices, err := Discover(config_g)
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        HttpLog(r, 500)
        log.Printf("Error: %s\n", err)
        return
    }
    if UpdateDevices(config_g, devices, newDevices) {
        log.Println("Device refresh needed, writing out to file")
        err = WriteDevices(config_g, devices)
        if err != nil {
            http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
            HttpLog(r, 500)
            log.Printf("Error: %s\n", err)
            return
        }
        message = "Device(s) updated"
    }

    tmpl, _ := template.ParseFiles(config_g.HtmlTemplate) // TODO: check for file first
    httpData := struct {
        Mode string
        DeviceData map[string]map[string]string
        NewDeviceData map[string]map[string]string
        Message string
    }{
        Mode: "discover",
        DeviceData: newDevices,
        Message: message,
    }

    err = tmpl.Execute(w, httpData)
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        HttpLog(r, 500)
        log.Printf("Error: %s\n", err)
        return
    }

    HttpLog(r, 200)
}



func apiListHandler(w http.ResponseWriter, r *http.Request) {
    devices, err := ReadDevices(config_g)
    if err != nil {
        log.Println(err)
        return
    }
    jsonOut, err := json.Marshal(devices)
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        HttpLog(r, 500)
        return
    }

    w.Write([]byte(jsonOut))
    HttpLog(r, 200)
}



func apiStatusHandler(w http.ResponseWriter, r *http.Request) {
    devices, err := ReadDevices(config_g)
    if err != nil {
        log.Println(err)
        return
    }

    rbody, err := ioutil.ReadAll(r.Body)
    if err != nil {
        log.Printf("ERROR: %s", err)
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        HttpLog(r, 500)
        return
    }

    var reqData map[string]string
    err = json.Unmarshal(rbody, &reqData)
    if err != nil {
        log.Printf("ERROR: %s", err)
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        HttpLog(r, 500)
        return
    }

    device := &wemo.Device{Host:devices[reqData["MacAddress"]]["Host"]}
    state := device.GetBinaryState()
    w.Write([]byte(fmt.Sprintf("%d", state)))
    HttpLog(r, 200)
}



func apiDiscoverHandler(w http.ResponseWriter, r *http.Request) {
    devices, err := ReadDevices(config_g)
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        HttpLog(r, 500)
        log.Printf("Error: %s\n", err)
        return
    }
    newDevices, err := Discover(config_g)
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        HttpLog(r, 500)
        log.Printf("Error: %s\n", err)
        return
    }
    if UpdateDevices(config_g, devices, newDevices) {
        log.Println("Device refresh needed, writing out to file")
        err = WriteDevices(config_g, devices)
        if err != nil {
            http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
            HttpLog(r, 500)
            log.Printf("Error: %s\n", err)
            return
        }
        w.Write([]byte("1"))
    } else {
        w.Write([]byte("0"))
    }

    HttpLog(r, 200)
}



func apiOnHandler(w http.ResponseWriter, r *http.Request) {
    devices, err := ReadDevices(config_g)
    if err != nil {
        log.Println(err)
        return
    }

    rbody, err := ioutil.ReadAll(r.Body)
    if err != nil {
        log.Printf("ERROR: %s", err)
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        HttpLog(r, 500)
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        return
    }

    var reqData map[string]string
    err = json.Unmarshal(rbody, &reqData)
    if err != nil {
        log.Printf("ERROR: %s", err)
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        HttpLog(r, 500)
        return
    }

    device := &wemo.Device{Host:devices[reqData["MacAddress"]]["Host"]}
    err = device.On()
    if err != nil {
        log.Printf("Could not turn on %s: %s", reqData["MacAddress"], err)
    } else {
        log.Printf("Turned on %s", reqData["MacAddress"])
    }

    HttpLog(r, 200)
}



func apiOffHandler(w http.ResponseWriter, r *http.Request) {
    devices, err := ReadDevices(config_g)
    if err != nil {
        log.Println(err)
        return
    }

    rbody, err := ioutil.ReadAll(r.Body)
    if err != nil {
        log.Printf("ERROR: %s", err)
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        HttpLog(r, 500)
        return
    }

    var reqData map[string]string
    err = json.Unmarshal(rbody, &reqData)
    if err != nil {
        log.Printf("ERROR: %s", err)
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        HttpLog(r, 500)
        return
    }

    device := &wemo.Device{Host:devices[reqData["MacAddress"]]["Host"]}
    err = device.Off()
    if err != nil {
        log.Printf("Could not turn off %s: %s", reqData["MacAddress"], err)
    } else {
        log.Printf("Turned off %s", reqData["MacAddress"])
    }

    HttpLog(r, 200)
}



func apiScheduleHandler(w http.ResponseWriter, r *http.Request) {
/*
    devices, err := ReadDevices(config_g)
    if err != nil {
        log.Println(err)
        return
    }

    rbody, err := ioutil.ReadAll(r.Body)
    if err != nil {
        log.Printf("ERROR: %s", err)
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        HttpLog(r, 500)
        return
    }
//*/
    readSchedule(config_g.ScheduleFile)
}



func apiHandler(w http.ResponseWriter, r *http.Request) {
}



func nullPage(w http.ResponseWriter, r *http.Request) {
}



func StartHttp (config Config_t) {
    config_g = config
    log.Printf("Starting webserver on port %d\n", config.HttpPort)

    http.HandleFunc("/", nullPage)

    // Web UI handlers
    http.HandleFunc("/ui", guiHandler)
    http.HandleFunc("/favicon.ico", iconHandler)
    http.HandleFunc("/discover", discoverHandler)

    // API handlers
    http.HandleFunc("/api/list", apiListHandler)
    http.HandleFunc("/api/status", apiStatusHandler)
    http.HandleFunc("/api/discover", apiDiscoverHandler)
    http.HandleFunc("/api/on", apiOnHandler)
    http.HandleFunc("/api/off", apiOffHandler)
    http.HandleFunc("/api/schedule", apiScheduleHandler)


    portStr := fmt.Sprintf(":%d", config.HttpPort)
    if config.UseTls == true {
        http.ListenAndServeTLS(portStr, config.TlsCertFile, config.TlsKeyFile, nil)
    } else {
        http.ListenAndServe(portStr, nil)
    }
}
