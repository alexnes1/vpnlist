package main

import (
	"fmt"
	"time"

	"github.com/go-ping/ping"
)

type VpnRecord struct {
	HostName       string
	IP             string
	Score          int
	Ping           int
	Speed          int
	CountryLong    string
	CountryShort   string
	NumVpnSessions int
	Uptime         int
	TotalUsers     int
	TotalTraffic   int64
	LogType        string
	Operator       string
	Message        string
	OpenVPNConfig  string
	Online         bool
	AvgPing        time.Duration
}

func (v *VpnRecord) String() string {
	return fmt.Sprintf("%-3s\t%-17s\t%-17s\t%-7.2f Mbps",
		v.CountryShort, v.IP, v.HostName, float32(v.Speed)/1000000)
}

func (v *VpnRecord) CheckPing(timeout time.Duration) {
	pinger, err := ping.NewPinger(v.IP)
	if err != nil {
		panic(err)
	}
	pinger.Count = 1
	pinger.Timeout = timeout
	pinger.Run()
	stats := pinger.Statistics()
	v.Online = stats.PacketsSent == stats.PacketsRecv
	v.AvgPing = stats.AvgRtt
}
