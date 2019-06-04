# wemoweb
Web interface for to control Wemo devices.
Work still in progress.

## Installation
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
