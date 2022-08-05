# vpnlist

A tiny manager of OpenVPN configs obtained from **vpngate.net** for Linux.

## Installation

Get a proper binary from the "Releases" page. Maybe make sure it's in your $PATH variable for convenience.

## Usage

### update

At first, you should get some servers from vpngate.net:
```
$ vpnlist update
```
vpngate.net API offers a fraction of its servers per request, so `vpnlist` stores them in the local sqlite database located in the `.config/vpnlist/` directory of the current user. 
The update command is meant to be used from time to time, so it may be a good idea to run it on schedule to incrementally expand the list of available servers.

### *default*
Show descriptions of all servers stored in the local database.
```
$ vpnlist
   	IP               	Host             	Speed       
ID 	xxx.xxx.xxx.xxx  	public-vpn-xxx   	22.42   Mbps
JP 	xxx.xxx.xxx.xxx  	public-vpn-xxx   	78.40   Mbps
...
```

Output of this command can be filtered with `country` and `speed` flags:
```
$ vpnlist --country JP -c US --speed 100
```

If you want to check the status of servers:
```
$ vpnlist --ping --ping-timeout 100ms --ping-workers 8
```

If `ping` flag is set, the default `ping-timeout` value is 500ms and `ping-workers` is 4.

Use `online` flag to display only online servers (this flag implies the same behaviour and defaults as `ping` flag).

vpnlist uses [go-ping](https://pkg.go.dev/github.com/sparrc/go-ping) library and therefore attempts to send an
"uprivileged" ping via UDP. If you want to ping servers, you should enable the following setting:
```
$ sudo sysctl -w net.ipv4.ping_group_range="0   2147483647"
```

### countries
Show the list of all countries represented in the local database.
```
$ vpnlist countries
Australia (AU)
Egypt (EG)
Indonesia (ID)
Japan (JP)
...
```

### show
Show config of a specific server
```
$ vpnlist show public-vpn-xxx
###############################################################################
# OpenVPN 2.0 Sample Configuration File
...
```

### random
Show a random config
```
$ vpnlist random
###############################################################################
# OpenVPN 2.0 Sample Configuration File
...
```
Random selection can be limited the same way as the defaul command:
```
$ vpnlist random --country JP --speed 20
```

## How to use with OpenVPN
1. Output of `show` and `random` commands can be stored in `.ovpn` file and used as usual.
2. Alternatively the output can be piped:
```
$ vpnlist random | sudo openvpn /dev/stdin
```


## Credits

* [VPN Gate](https://www.vpngate.net/en/) - Public Free VPN Cloud by Univ of Tsukuba, Japan
