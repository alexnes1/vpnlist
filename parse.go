package main

import (
	"bufio"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const apiURL = "http://www.vpngate.net/api/iphone/"

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
}

func (v VpnRecord) Filename() string {
	return fmt.Sprintf("%s_%s_%s.ovpn", v.CountryShort, v.HostName, v.IP)
}

func (v VpnRecord) String() string {
	return fmt.Sprintf("%-3s\t%-17s\t%-17s\t%-7.2f Mbps",
		v.CountryShort, v.IP, v.HostName, float32(v.Speed)/1000000)
}

func makeRecordFromCsvRow(row []string) (VpnRecord, error) {
	rec := VpnRecord{}
	for i, part := range row {
		switch i {
		case 0:
			rec.HostName = part
		case 1:
			rec.IP = part
		case 2:
			val, err := strconv.Atoi(part)
			if err != nil {
				return rec, err
			}
			rec.Score = val
		case 3:
			val, err := strconv.Atoi(part)
			if err != nil {
				return rec, err
			}
			rec.Ping = val
		case 4:
			val, err := strconv.Atoi(part)
			if err != nil {
				return rec, err
			}
			rec.Speed = val
		case 5:
			rec.CountryLong = part
		case 6:
			rec.CountryShort = part
		case 7:
			val, err := strconv.Atoi(part)
			if err != nil {
				return rec, err
			}
			rec.NumVpnSessions = val
		case 8:
			val, err := strconv.Atoi(part)
			if err != nil {
				return rec, err
			}
			rec.Uptime = val
		case 9:
			val, err := strconv.Atoi(part)
			if err != nil {
				return rec, err
			}
			rec.TotalUsers = val
		case 10:
			val, err := strconv.ParseInt(part, 10, 64)
			if err != nil {
				return rec, err
			}
			rec.TotalTraffic = val
		case 11:
			rec.LogType = part
		case 12:
			rec.Operator = part
		case 13:
			rec.Message = part
		case 14:
			val, err := base64.StdEncoding.DecodeString(part)
			if err != nil {
				return rec, err
			}
			rec.OpenVPNConfig = string(val)
		default:
			return rec, fmt.Errorf("unexpected column")
		}
	}

	return rec, nil
}

func downloadRecords(output io.Writer) ([]VpnRecord, error) {
	resp, err := http.Get(apiURL)
	if err != nil {
		return []VpnRecord{}, err
	}

	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)

	for i := 0; i < 2; i++ {
		_, _, err = reader.ReadLine()
		if err != nil {
			return []VpnRecord{}, err
		}
	}

	vpnRecords := []VpnRecord{}

	csvReader := csv.NewReader(reader)

	for {
		row, err := csvReader.Read()
		if row == nil && err == io.EOF {
			break
		}
		if err != nil && row != nil {
			continue
		}
		if err != nil && row == nil {
			return []VpnRecord{}, err
		}

		record, err := makeRecordFromCsvRow(row)
		if err != nil {
			continue
		}
		vpnRecords = append(vpnRecords, record)
		fmt.Fprintf(output, "%s\n", record)
	}

	return vpnRecords, nil
}
