package wemoweb

import (
    "io/ioutil"
    "gopkg.in/yaml.v2"
    "fmt"
    "strings"
    "strconv"
    "log"
    "sort"
    "time"
)

type Schedule struct {
    Mac string
    Timeline map[int]map[int]string
    SortedTimes []string
}



func readSchedule(filename string) (map[string]Schedule, error) {
    scheduleData, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    var scheduleYaml map[string]map[string]map[string][]map[string]string
    err = yaml.Unmarshal(scheduleData, &scheduleYaml)


    schedule := make(map[string]Schedule)
    for mac, v := range scheduleYaml["schedule"] {
        var s Schedule
        //fmt.Println("MAC:", mac)

        s.Mac = mac
        s.Timeline = make(map[int]map[int]string)
        for _, state := range []string{ "on", "off" } {
            for _, event := range v[state] {
                time := strings.Split(event["time"], ":")
                hour, err := strconv.Atoi(time[0])
                if err != nil {
                    log.Fatal(err)
                }
                minute, err := strconv.Atoi(time[1])
                if err != nil {
                    log.Fatal(err)
                }
                if s.Timeline[hour] == nil {
                    s.Timeline[hour] = make(map[int]string)
                }
                s.Timeline[hour][minute] = state
                s.SortedTimes = append(s.SortedTimes, fmt.Sprintf("%02d%02d", hour, minute))
            }
        }
        sort.Strings(s.SortedTimes)
        schedule[mac] = s
    }

    return schedule, nil
}



func (s Schedule) GetCurrentStatus() (string) {
    now := time.Now()

    nowString := fmt.Sprintf("%02d%02d", now.Hour(), now.Minute())

    eventTime := ""
    for i := range s.SortedTimes {
        r := strings.Compare(nowString, s.SortedTimes[i])
        if r == 0 {
            eventTime = s.SortedTimes[i]
            break
        } else if r == -1 {
            if i == 0 {
                eventTime = s.SortedTimes[len(s.SortedTimes)-1]
                break
            } else {
                eventTime = s.SortedTimes[i-1]
                break
            }
        }
    }
    fmt.Printf("eventTime: %s\n", eventTime)
    timeParts := strings.Split(eventTime, "")
    hour, err := strconv.Atoi(timeParts[0]+timeParts[1])
    if err != nil {
        log.Fatal(err)
    }
    minute, err := strconv.Atoi(timeParts[2]+timeParts[3])
    if err != nil {
        log.Fatal(err)
    }

    return s.Timeline[hour][minute]
}



func ShowScheduleCli(config Config_t) (error) {
    devices, err := ReadDevices(config)
    schedule, err := readSchedule(config.ScheduleFile)
    if err != nil {
        return err
    }

    for mac, s := range schedule {
        fmt.Printf("%s should be '%s'\n", devices[mac]["FriendlyName"], s.GetCurrentStatus())
    }


    return nil
}
