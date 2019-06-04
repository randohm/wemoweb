# wemoweb
A simple web interface to control Wemo® devices.
Work still in progress.
This is done with the goal of being used on mobile device web browsers, as well as larger devices.


This may not work for everyone.
I keep my devices' IPs static and that makes it easier and require fewer discoveries.
It should work if the devices do not change IPs often.

## Compiling
No Makefile yet.
Just run: `go build wemoweb.go`

## Use

### Discovery

Run: `./wemoweb discover`

It will scan and prompt to save a devices.json file.

### Webserver

Run: `/.wemoweb server`


## Application Installation
Create `config.json` out of `sample.config.json`.

**Config File Sample**  
```
{
  "http_port": 8080,
  "eth_device": "eth0",
  "devices_file": "devices.json",
  "discovery_timeout": 5,
  "html_tmpl": "index.html.tpl",
  "users_file": "users.json",
  "use_tls": true,
  "tls_cert_file": "cert.pem",
  "tls_key_file": "key.pem"
}
```

### Confing File Fields
- http\_port: TCP port daemon binds to for listens.
- eth\_device: Ethernet device on the same subnet as the Wemo® devices.
- devices\_json: JSON file with list of Wemo® devices.
- discovery\_timeout": Timeout in seconds for device scans.
- html\_tmpl: HTML template file. Start with the provided index.html.tpl.
- users\_file: JSON file containing user credentials. Not having this value in the config file disables user authentication.
- use\_tls: Boolean for whether the daemon uses TLS on plaintext.
- tls\_cert\_file: PEM file containing the TLS cert.
- tls\_key\_file: PEM file containing the private key.


### User File
Format:
```
{
    "username": "MD5 sum of password"
}
```

### Devices File
Format:
```
{
  "Device 1": {
    "ip_port": "192.168.0.10:49153"
  },
  "Device 2": {
    "ip_port": "192.168.0.11:49153"
  }
}
```

