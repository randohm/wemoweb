package wemoweb

import (
    "net/http"
    "html/template"
    golog "log"
    "fmt"
    "crypto/md5"
    "github.com/randohm/go.wemo"
    "io/ioutil"
    "encoding/json"
    "gopkg.in/yaml.v2"
    "strconv"
    "time"
    "os"
    "sort"
)



var gologger *golog.Logger
var users map[string]string



/* Reads users file. 
   Returns:
     map with usernames as keys, md5sum of password
     error
*/
func readUsers() (map[string]string, error) {
    log.Tracef("Reading in users file")
    usersJson, err := ioutil.ReadFile(config.UsersFile)
    if err != nil {
        log.Errorf("%s", err)
        return map[string]string{}, err
    }

    var users map[string]map[string]string
    err = yaml.Unmarshal(usersJson, &users)
    if err != nil {
        log.Errorf("%s", err)
        return nil, err
    }
    return users["users"], nil
}



/*
    Uses Go's log to print out a standard webserver access log.
*/
func httpLog(r *http.Request, status int) {
    user, _, _ := r.BasicAuth()
    if user == "" {
        user = "-"
    }
    gologger.Printf("%s %s %d %s %s %s (%s)", r.RemoteAddr, user, status, r.Method, r.URL, r.Proto, r.UserAgent())
}



func checkUserPass(user, pass string) bool {
    if users == nil {
        var err error
        users, err = readUsers()
        if err != nil {
            log.Errorf("%s", err)
            return false
        }
    }
    passMd5 := fmt.Sprintf("%x", md5.Sum([]byte(pass)))
    if users[user] == passMd5 {
        return true
    }

    return false
}



func checkHttpAuth(w http.ResponseWriter, r *http.Request) bool {
    user, pass, ok := r.BasicAuth()
    if !ok || !checkUserPass(user, pass) {
        http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
        httpLog(r, http.StatusUnauthorized)
        return false
    }
    return true
}



func generateRootPage(w http.ResponseWriter, devices map[string]map[string]string, message string) error {
    var devicesList []map[string]string
    for mac, v := range devices {
        d := &wemo.Device{Host:v["Host"]}
        v["state"] = fmt.Sprintf("%d", d.GetBinaryState())
        devicesList = append(devicesList, map[string]string{ "Mac": mac, "FriendlyName": v["FriendlyName"], "state": v["state"] })
    }
    sort.Slice(devicesList, func(i, j int) bool { return devicesList[i]["FriendlyName"] < devicesList[j]["FriendlyName"]})
    log.Tracef("DevicesList: %+v", devicesList)

    _, err := os.Stat(config.HtmlTemplate)
    if err != nil {
        log.Errorf("%s", err)
        return err
    }

    tmpl, _ := template.ParseFiles(config.HtmlTemplate)
    httpData := struct {
        Mode string
        DeviceData []map[string]string
        Message string
    }{
        Mode: "main",
        DeviceData: devicesList,
        Message: message,
    }
    log.Tracef("httpData: %+v", httpData)
    err = tmpl.Execute(w, httpData)
    if err != nil {
        log.Errorf("%s", err)
        return err
    }

    return nil
}



func guiHandler(w http.ResponseWriter, r *http.Request) {
    if config.UsersFile != "" {
        if !checkHttpAuth(w, r) {
            return
        }
    }

    var message string
    devices, err := readDevices()
    if err != nil {
        log.Errorf("%s", err)
        return
    }
    op, _ := r.URL.Query()["op"]
    dev, _ := r.URL.Query()["dev"]
    length, _ := r.URL.Query()["len"]
    if len(op) > 0 && len(dev) > 0 {
        device := &wemo.Device{Host:devices[dev[0]]["Host"]}
        switch op[0] {
            case "on":
                log.Debugf("Turning on %s", devices[dev[0]]["FriendlyName"])
                err = device.On()
                if err != nil {
                    message = fmt.Sprintf("Could not turn on %s: %s", devices[dev[0]]["FriendlyName"], err)
                } else {
                    message = fmt.Sprintf("Turned on %s", devices[dev[0]]["FriendlyName"])
                }
            case "off":
                log.Debugf("Turning off %s", devices[dev[0]]["FriendlyName"])
                err = device.Off()
                if err != nil {
                    message = fmt.Sprintf("Could not turn off %s: %s", devices[dev[0]]["FriendlyName"], err)
                } else {
                    message = fmt.Sprintf("Turned off %s", devices[dev[0]]["FriendlyName"])
                }
            case "timer":
                if len(length) > 0 {
                    log.Debugf("Turning on %s for %s minutes", devices[dev[0]]["FriendlyName"], length[0])
                    minutes, _ := strconv.Atoi(length[0])
                    err = device.On()
                    if err != nil {
                        message = fmt.Sprintf("Could not turn on %s: %s", devices[dev[0]]["FriendlyName"], err)
                    } else {
                        message = fmt.Sprintf("Turned on %s", devices[dev[0]]["FriendlyName"])
                    }
                    go timerOff(device, minutes)
                    message = fmt.Sprintf("Set timer on %s for %s minutes", devices[dev[0]]["FriendlyName"], length[0])
                } else {
                    http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
                    httpLog(r, http.StatusInternalServerError)
                    log.Errorf("Length not specified")
                    return
                }
        }
    }
    err = generateRootPage(w, devices, message)
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        httpLog(r, http.StatusInternalServerError)
        log.Errorf("%s", err)
        return
    }

    httpLog(r, http.StatusOK)
}



func timerOff(device *wemo.Device, minutes int) {
    sleepTime := time.Duration(minutes) * time.Minute
    log.Debugf("Setting timer for %+v", sleepTime)
    time.Sleep(sleepTime)
    err := device.Off()
    if err != nil {
        log.Errorf("Error turning off after sleep: %s", err)
        return
    }
    log.Debugf("Turned off after %+v", sleepTime)
}



func iconHandler(w http.ResponseWriter, r *http.Request) {
    _, err := os.Stat(config.FavIcon)
    if err != nil {
        log.Infof("favicon file '%s' not found: %s", config.FavIcon, err)
        http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
        httpLog(r, http.StatusNotFound)
        return
    }

    iconData, err := ioutil.ReadFile(config.FavIcon)
    if err != nil {
        log.Errorf("Could not load favicon file '%s': %s", config.FavIcon, err)
        http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
        httpLog(r, http.StatusNotFound)
        return
    }

    w.Write(iconData)
    httpLog(r, http.StatusOK)
}



func discoverHandler(w http.ResponseWriter, r *http.Request) {
    message := "No detected device changes"
    devices, err := readDevices()
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        httpLog(r, http.StatusInternalServerError)
        log.Errorf("%s", err)
        return
    }

    newDevices, err := discover()
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        httpLog(r, http.StatusInternalServerError)
        log.Errorf("%s", err)
        return
    }
    if updateDevices(devices, newDevices) {
        log.Debugf("Device refresh needed, writing out to file")
        err = writeDevices(devices)
        if err != nil {
            http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
            httpLog(r, http.StatusInternalServerError)
            log.Errorf("%s", err)
            return
        }
        message = "Device(s) updated"
    }

    // Create device list sorted by FriendlyName
    var newDevicesList []map[string]string
    for _, v := range newDevices {
        newDevicesList = append(newDevicesList, v)
    }
    sort.Slice(newDevicesList, func(i, j int) bool { return newDevicesList[i]["FriendlyName"] < newDevicesList[j]["FriendlyName"]})

    _, err = os.Stat(config.HtmlTemplate)
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        httpLog(r, http.StatusInternalServerError)
        log.Errorf("%s", err)
        return
    }
    tmpl, _ := template.ParseFiles(config.HtmlTemplate)
    httpData := struct {
        Mode string
        DeviceData []map[string]string
        Message string
    }{
        Mode: "discover",
        DeviceData: newDevicesList,
        Message: message,
    }

    err = tmpl.Execute(w, httpData)
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        httpLog(r, http.StatusInternalServerError)
        log.Errorf("%s", err)
        return
    }

    httpLog(r, http.StatusOK)
}



func scheduleHandler(w http.ResponseWriter, r *http.Request) {
    message := ""

    devices, err := readDevices()
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        httpLog(r, http.StatusInternalServerError)
        log.Errorf("%s", err)
        return
    }

    schedule, err := readSchedule(config.ScheduleFile)
    if err != nil {
        log.Errorf("%s", err)
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        httpLog(r, http.StatusInternalServerError)
        return
    }

    _, err = os.Stat(config.HtmlTemplate)
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        httpLog(r, http.StatusInternalServerError)
        log.Errorf("%s", err)
        return
    }

    var scheduleList []ScheduleItem
    for mac, v := range schedule {
        v.FriendlyName = devices[mac]["FriendlyName"]
        scheduleList = append(scheduleList, v)
    }
    sort.Slice(scheduleList, func(i, j int) bool { return scheduleList[i].FriendlyName < scheduleList[j].FriendlyName})

    tmpl, _ := template.ParseFiles(config.HtmlTemplate)
    httpData := struct {
        Mode string
        ScheduleData []ScheduleItem
        Message string
    }{
        Mode: "schedule",
        ScheduleData: scheduleList,
        Message: message,
    }

    err = tmpl.Execute(w, httpData)
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        httpLog(r, http.StatusInternalServerError)
        log.Errorf("%s", err)
        return
    }
    httpLog(r, http.StatusOK)
}



func apiListHandler(w http.ResponseWriter, r *http.Request) {
    devices, err := readDevices()
    if err != nil {
        log.Errorf("%s", err)
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        httpLog(r, http.StatusInternalServerError)
        return
    }

    jsonOut, err := json.Marshal(devices)
    if err != nil {
        log.Errorf("%s", err)
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        httpLog(r, http.StatusInternalServerError)
        return
    }
    log.Tracef("Devices Json: %s", jsonOut)

    w.Write([]byte(jsonOut))
    httpLog(r, http.StatusOK)
}



func apiStatusHandler(w http.ResponseWriter, r *http.Request) {
    devices, err := readDevices()
    if err != nil {
        log.Errorf("%s", err)
        return
    }

    rbody, err := ioutil.ReadAll(r.Body)
    if err != nil {
        log.Errorf("%s", err)
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        httpLog(r, http.StatusInternalServerError)
        return
    }
    log.Tracef("Request Body: '%s'", rbody)

    if len(rbody) == 0 {
        log.Debugf("Empty request body")
        http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
        httpLog(r, http.StatusBadRequest)
        return
    }

    var reqData map[string]string
    err = json.Unmarshal(rbody, &reqData)
    if err != nil {
        log.Errorf("%s", err)
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        httpLog(r, http.StatusInternalServerError)
        return
    }
    log.Tracef("Request Json: %+v", reqData)

    device := &wemo.Device{Host:devices[reqData["MacAddress"]]["Host"]}
    state := "off"
    if device.GetBinaryState() == 1 {
        state = "on"
    }
    outData := map[string]string{"state": state}
    jsonOut, err := json.Marshal(outData)
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        httpLog(r, http.StatusInternalServerError)
        return
    }
    w.Write([]byte(jsonOut))
    httpLog(r, http.StatusOK)
}



func apiDiscoverHandler(w http.ResponseWriter, r *http.Request) {
    devices, err := readDevices()
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        httpLog(r, http.StatusInternalServerError)
        log.Errorf("%s", err)
        return
    }

    newDevices, err := discover()
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        httpLog(r, http.StatusInternalServerError)
        log.Errorf("%s", err)
        return
    }
    log.Tracef("Discovered devices: %+v", newDevices)

    devicesUpdated := updateDevices(devices, newDevices)
    if devicesUpdated {
        log.Debugf("Device refresh needed, writing out to file")
        err = writeDevices(devices)
        if err != nil {
            http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
            httpLog(r, http.StatusInternalServerError)
            log.Errorf("%s", err)
            return
        }
    } else {
        log.Debugf("No device changes found")
    }

    outData := map[string]bool{"updated": devicesUpdated}
    jsonOut, err := json.Marshal(outData)
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        httpLog(r, http.StatusInternalServerError)
        return
    }
    w.Write([]byte(jsonOut))
    httpLog(r, http.StatusOK)
}



func apiOnHandler(w http.ResponseWriter, r *http.Request) {
    devices, err := readDevices()
    if err != nil {
        log.Errorf("%s", err)
        return
    }

    rbody, err := ioutil.ReadAll(r.Body)
    if err != nil {
        log.Errorf("%s", err)
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        httpLog(r, http.StatusInternalServerError)
        return
    }
    log.Tracef("Request Body: '%s'", rbody)

    var reqData map[string]string
    err = json.Unmarshal(rbody, &reqData)
    if err != nil {
        log.Errorf("%s", err)
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        httpLog(r, http.StatusInternalServerError)
        return
    }
    log.Tracef("Request Json: %+v", reqData)

    device := &wemo.Device{Host:devices[reqData["MacAddress"]]["Host"]}
    err = device.On()
    if err != nil {
        log.Warningf("Could not turn on %s: %s", reqData["MacAddress"], err)
    } else {
        log.Debugf("Turned on %s", reqData["MacAddress"])
    }

    httpLog(r, http.StatusOK)
}



func apiOffHandler(w http.ResponseWriter, r *http.Request) {
    devices, err := readDevices()
    if err != nil {
        log.Errorf("%s", err)
        return
    }

    rbody, err := ioutil.ReadAll(r.Body)
    if err != nil {
        log.Errorf("%s", err)
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        httpLog(r, http.StatusInternalServerError)
        return
    }
    log.Tracef("Request Body: '%s'", rbody)

    var reqData map[string]string
    err = json.Unmarshal(rbody, &reqData)
    if err != nil {
        log.Errorf("%s", err)
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        httpLog(r, http.StatusInternalServerError)
        return
    }
    log.Tracef("Request Json: %+v", reqData)

    device := &wemo.Device{Host:devices[reqData["MacAddress"]]["Host"]}
    err = device.Off()
    if err != nil {
        log.Warningf("Could not turn off %s: %s", reqData["MacAddress"], err)
    } else {
        log.Debugf("Turned off %s", reqData["MacAddress"])
    }

    httpLog(r, http.StatusOK)
}



func apiScheduleHandler(w http.ResponseWriter, r *http.Request) {
    schedule, err := readSchedule(config.ScheduleFile)
    if err != nil {
        log.Errorf("%s", err)
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        httpLog(r, http.StatusInternalServerError)
        return
    }

    jsonOut, err := json.Marshal(schedule)
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        httpLog(r, http.StatusInternalServerError)
        return
    }
    w.Write([]byte(jsonOut))

    httpLog(r, http.StatusOK)
}



func apiHandler(w http.ResponseWriter, r *http.Request) {
}



func nullPage(w http.ResponseWriter, r *http.Request) {
}



/*
    Assigns handlers to URLs, starts the HTTP server
*/
func StartHttp () {
    gologger = golog.New(golog.Writer(), golog.Prefix(), golog.Flags())
    log.Infof("Starting webserver on %s", config.Listen)

    http.HandleFunc("/", nullPage)

    // Web UI handlers
    http.HandleFunc("/ui", guiHandler)
    http.HandleFunc("/ui/discover", discoverHandler)
    http.HandleFunc("/ui/schedule", scheduleHandler)
    http.HandleFunc("/favicon.ico", iconHandler)

    // API handlers
    http.HandleFunc("/api/list", apiListHandler)
    http.HandleFunc("/api/status", apiStatusHandler)
    http.HandleFunc("/api/discover", apiDiscoverHandler)
    http.HandleFunc("/api/on", apiOnHandler)
    http.HandleFunc("/api/off", apiOffHandler)
    http.HandleFunc("/api/schedule", apiScheduleHandler)

    var err error
    if config.UseTls == true {
        err = http.ListenAndServeTLS(config.Listen, config.TlsCertFile, config.TlsKeyFile, nil)
    } else {
        err = http.ListenAndServe(config.Listen, nil)
    }
    if err != nil {
        log.Errorf("Could not start webserver: %s", err)
        os.Exit(1)
    }
}
