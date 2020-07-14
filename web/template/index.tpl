<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>WemoWeb</title>
  <style TYPE="text/css">
<!--
body {
  font-family: Arial;
  font-size: 18px;
  margin: 0px 0px 0px 0px;
  padding: 0px 0px 0px 0px;
}

table.device_table {
  border-spacing: 0px 1px;
  font-size: 20px;
  border: 0px solid black;
  padding: 0px 0px 0px 0px;
  width: 500px;
}

tr.device_tr_active {
  background: #00aa00;
  border-spacing: 1px 0px;
}

tr.device_tr_inactive {
  background: #aaaaaa;
}

tr.discover_tr {
  background: #cccccc;
}

td.device_td_h1 {
  padding: 5px;
  width: 60%;
  text-align: center;
  font-size: 28px;
}

td.device_td_h2 {
  padding: 5px;
  width: 60%;
  text-align: left;
  font-size: 22px;
}

td.device_td_name {
  padding: 5px;
  width: 60%;
  text-align: left;
}

td.device_td_button {
  padding: 5px;
  text-align: right;
}

td.td_notfound {
  padding: 5px;
  text-align: center;
  font-size: 12px;
}

button.action_button {
  font-size: 18px;
}

button.timer_button {
  font-size: 18px;
}

input.minute_input {
  font-size: 16px;
  size: 2;
}


@media only screen and (max-width: 1080px) {
  body {
    font-family: Arial;
    font-size: 6vw;
    margin: 0px 0px 0px 0px;
    padding: 0px 0px 0px 0px;
  }

  table.device_table {
    border-spacing: 0px;
    font-size: 8vw;
    width: 100%;
    border: 0px;
    padding: 0px;
  }

  td.device_td {
    padding: 10px;
  }

  button.action_button {
    font-size: 5vw;
  }

  button.timer_button {
    font-size: 3vw;
  }

  input.minute_input {
    font-size: 3vw;
    size: 2;
  }

}
-->
  </style>
</head>

<body>

<table class="device_table">
<tbody>
{{ if eq .Mode "main" -}}
  {{- range .DeviceData }}
<tr class="{{if eq .state "1"}}device_tr_active{{else}}device_tr_inactive{{end}}">
    <td class="device_td_name">{{.FriendlyName}}</td>
    {{- if not (eq .state "-1")}}
    <td class="device_td_button">
        <button class="action_button" OnClick="window.location.href='/ui?op={{if eq .state "1"}}off{{else}}on{{end}}&dev={{.Mac}}'">{{if eq .state "1"}}Off{{else}}On{{end}}</button>
    </td>
    <td class="device_td_button">
        <form>
        <input type="hidden" name="op" value="timer"/>
        <input type="hidden" name="dev" value="{{.Mac}}"/>
        <input type="text" class="minute_input" value="15" size="2" maxlength="2" name="len"/>
        <button class="timer_button" OnClick="form.submit()">T</button>
        </form>
    </td>
    {{- else }}
    <td colspan="2" class="td_notfound">Not found</td>
    {{- end }}
</tr>
  {{- end }}
{{ else if eq .Mode "discover" -}}
<tr>
    <td class="device_td_h1" colspan="2">Discovering Devices</td>
</tr>
<tr><th>Discovered Device</th><th>IP:port</th></tr>
  {{- range .DeviceData }}
<tr class="discover_tr">
    <td class="device_td">{{.FriendlyName}}</td><td class="device_td">{{.Host}}</td>
</tr>
  {{- end }}
{{ else if eq .Mode "schedule" -}}
<tr>
    <td class="device_td_h1" colspan="2">Schedules</td>
</tr>
  {{- range .ScheduleData }}
<tr>
    <td class="device_td_h2">{{.FriendlyName}}</td>
</tr>
    {{- range .Timeline  }}
<tr class="{{if eq .state "on"}}device_tr_active{{else}}device_tr_inactive{{end}}">
    <td class="device_td_name">{{.time}}</td>
    <td class="device_td_name">{{.state}}</td>
</tr>
    {{- end }}
  {{- end }}
{{ end -}}
</tbody>
</table>

<p>{{.Message}}</p>

<table class="device_table">
<tbody>
<tr>
    <td style="text-align:center"><a href="/ui">{{ if eq .Mode "main" -}}Refresh{{else -}}Home{{end -}}</a></td>
    <td style="text-align:center"><a href="/ui/schedule">Schedule</a></td>
    <td style="text-align:center"><a href="/ui/discover">Discover</a></td>
</tr>
</tbody>
</table>

</body>
</html>
