<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>Wemo Control</title>
  <style TYPE="text/css">
<!--
body {
  font-family: Arial;
  font-size: 18px;
  margin: 0px 0px 0px 0px;
  padding: 0px 0px 0px 0px;
}

table.device_table {
  border-spacing: 0px;
  font-size: 20px;
  border: 0px;
  padding: 0px;
  width: 500px;
}

td.device_td {
  padding: 5px;
}

button.action_button {
  font-size: 18px;
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
    font-size: 6vw;
  }

}
-->
  </style>
</head>

<body>

<table class="device_table">
<tbody>
{{range $key, $value := .DeviceData -}}
<tr bgcolor="{{if eq $value.state "1"}}#00aa00{{else}}#aaaaaa{{end}}">
    <td class="device_td">{{$key}}</td><td class="device_td"><button class="action_button" OnClick="window.location.href='/?op={{if eq $value.state "1"}}off{{else}}on{{end}}&dev={{$key}}'">{{if eq $value.state "1"}}Off{{else}}On{{end}}</button></td>
</tr>
{{end -}}
</tbody>
</table>

<a href="/">Refresh</a>

</body>
</html>
