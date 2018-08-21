# Control Interface for wpa_supplicant in Go

[![ISC License](https://img.shields.io/badge/license-ISC-blue.svg)](https://github.com/dangerousHobo/wpasupplicant/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/dangerousHobo/wpasupplicant)](https://goreportcard.com/report/github.com/dangerousHobo/wpasupplicant)

Package wpasupplicant provides a control interface to a wpa_supplicant process.

The connection to the wpa_supplicant process is by a Unix socket. The control
interface is defined in your wpa_supplicant.conf file.

Example of a wpa_supplicant.conf:

```
ctrl_interface=/var/run/wpa_supplicant
```

To open a connection:

```go
uconn, err := wpasupplicant.Connect("/tmp/our-socket", "/var/run/wpa_supplicant")
```

From this point you can start configuring for your network:

```go
usock.SetNetworkQuoted(id, "ssid", "foo")
usock.SetNetworkQuoted(id, "psk", "bar")
usock.SetNetwork(id, "proto", "WPA2")
usock.SetNetwork(id, "key_mgmt", "WPA-PSK")
```

How to know when to use SetNetwork vs SetNetworkQuoted? Read the wpa_supplicant.conf
documentation.

https://w1.fi/cgit/hostap/plain/wpa_supplicant/wpa_supplicant.conf

For further information on the wpa_supplicant control interface:

http://w1.fi/wpa_supplicant/devel/ctrl_iface_page.html
