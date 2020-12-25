package wemoweb

import (
    "io/ioutil"
    "gopkg.in/yaml.v2"
    "fmt"
    "strings"
    "strconv"
    "sort"
    "time"
    "github.com/randohm/go.wemo"
)



/*
    Represents a schedule for one device.

    Timeline is an array of maps with keys 'time' and 'state', sorted by time.
    'time' is a 4-digit string representation of 24-hour time of a schedule event in the format 'HHMM', zero-padded, no colon.
*/
type ScheduleItem struct {
    Mac string // MAC address, no delimiters
    Timeline []map[string]string // Sorted array of maps 
    FriendlyName string     // Should only be used for HTML templating
}



/*
    Reads in the schedule YAML file.
    Creates ScheduleItem structs for each device in the schedule

    Args:
      string: filename of schedule file

    Returns:
      map[string]ScheduleItem: key is the device MAC address
      error
*/
func readSchedule(filename string) (map[string]ScheduleItem, error) {
    scheduleData, err := ioutil.ReadFile(filename)
    if err != nil {
        log.Errorf("%s", err)
        return nil, err
    }
    log.Tracef("Schedule text:\n%s", scheduleData)

    var scheduleYaml map[string]map[string]map[string][]map[string]string
    err = yaml.Unmarshal(scheduleData, &scheduleYaml)

    schedule := make(map[string]ScheduleItem)
    if sched, ok := scheduleYaml["schedule"]; ok {
        for mac, v := range sched {
            var s ScheduleItem

            s.Mac = mac
            for _, state := range []string{ "on", "off" } {
                for _, event := range v[state] {
                    // Enforce 2-digit, zero-padded integers
                    time := strings.Split(event["time"], ":")
                    hour, err := strconv.Atoi(time[0])
                    if err != nil {
                        log.Errorf("%s", err)
                        return nil, err
                    }
                    minute, err := strconv.Atoi(time[1])
                    if err != nil {
                        log.Errorf("%s", err)
                        return nil, err
                    }
                    eventTime := fmt.Sprintf("%02d%02d", hour, minute)
                    s.Timeline = append(s.Timeline, map[string]string{ "time": eventTime, "state": state})
                }
            }
            sort.Slice(s.Timeline, func(i, j int) bool { return s.Timeline[i]["time"] < s.Timeline[j]["time"]})
            schedule[mac] = s
        }
    }

    log.Tracef("Schedule: %+v", schedule)
    return schedule, nil
}



/*
    Finds the most recently passed time in the schedule.
    Returns the state of that time.

    Returns:
      string: "on" or "off"
      error
*/
func (s ScheduleItem) GetScheduledState() (string, error) {
    now := time.Now()

    nowString := fmt.Sprintf("%02d%02d", now.Hour(), now.Minute())

    if len(s.Timeline) == 0 {
        return "undefined", nil
    }

    foundIdx := -1
    for i := range s.Timeline {
        r := strings.Compare(nowString, s.Timeline[i]["time"])
        if r == 0 {
            foundIdx = i
            break
        } else if r == -1 {
            if i == 0 {
                foundIdx = len(s.Timeline)-1
                break
            } else {
                foundIdx = i-1
                break
            }
        }
    }

    if foundIdx == -1 {
        foundIdx = len(s.Timeline)-1
    }
    log.Debugf("Last eventTime: %s %s\n", s.Mac, s.Timeline[foundIdx]["time"])

    return s.Timeline[foundIdx]["state"], nil
}



func getEventTime(eventTime string) (hour int, minute int, err error) {
    timeParts := strings.Split(eventTime, "")
    hour, err = strconv.Atoi(timeParts[0]+timeParts[1])
    if err != nil {
        log.Errorf("%s", err)
        return -1, -1, err
    }
    minute, err = strconv.Atoi(timeParts[2]+timeParts[3])
    if err != nil {
        log.Errorf("%s", err)
        return -1, -1, err
    }

    return hour, minute, nil
}



/*
    Prints each devices schedule with times sorted to stdout.

    Returns:
      error
*/
func ShowScheduleCli() (error) {
    devices, err := readDevices()
    if err != nil {
        log.Errorf("%s", err)
        return err
    }
    schedule, err := readSchedule(config.ScheduleFile)
    if err != nil {
        log.Errorf("%s", err)
        return err
    }

    for mac, s := range schedule {
        state, err := s.GetScheduledState()
        if err != nil {
            log.Errorf("%s", err)
            return err
        }
        fmt.Printf("%s should be '%s'\n", devices[mac]["FriendlyName"], state)
        for i := range s.Timeline {
            time := strings.Split(s.Timeline[i]["time"], "")
            if err != nil {
                log.Errorf("%s", err)
                return err
            }
            fmt.Printf("%s%s:%s%s %-4s\n", time[0], time[1], time[2], time[3], s.Timeline[i]["state"])
        }
        fmt.Printf("\n")
    }

    return nil
}



/*
    Checks if each device's schedule is being observed.
    Turns on/off the device if needed
*/
func EnforceSchedule() (error) {
    devices, err := readDevices()
    if err != nil {
        log.Errorf("%s", err)
        return err
    }
    schedule, err := readSchedule(config.ScheduleFile)
    if err != nil {
        log.Errorf("%s", err)
        return err
    }

    for mac, s := range schedule {
        d := &wemo.Device{Host:devices[mac]["Host"]}
        currentState := "off"
        if d.GetBinaryState() == 1 {
            currentState = "on"
        }
        scheduledState, err := s.GetScheduledState()
        if err != nil {
            log.Errorf("%s", err)
            return err
        }
        log.Debugf("%s is '%s' should be '%s'\n", devices[mac]["FriendlyName"], currentState, scheduledState)

        if currentState != scheduledState {
            device := &wemo.Device{Host:devices[mac]["Host"]}
            if scheduledState == "on" {
                err = device.On()
            } else if scheduledState == "off" {
                err = device.Off()
            }
            if err != nil {
                log.Warningf("Could not turn on %s: %s", devices[mac]["FriendlyName"], err)
            } else {
                log.Infof("Turned %s %s", scheduledState, devices[mac]["FriendlyName"])
            }
        }
    }

    return nil
}



func saveSchedule() (error) {
    return nil
}



func RunScheduler() {
    for {
        log.Debugf("Scheduler enforcing states")
        EnforceSchedule()
        time.Sleep(time.Duration(1) * time.Minute)
    }
}
