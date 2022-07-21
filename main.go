package main

import (
	"bufio"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
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
	OpenVPNConfig  []byte
}

func (v VpnRecord) Filename() string {
	return fmt.Sprintf("%s_%s_%s.ovpn", v.CountryShort, v.HostName, v.IP)
}

func (v VpnRecord) String() string {
	return fmt.Sprintf("%s\t%s\t%s\t%d\t%d", v.CountryLong, v.IP, v.HostName, v.Speed, v.Ping)
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
			rec.OpenVPNConfig = val
		default:
			return rec, fmt.Errorf("unexpected column")
		}
	}

	return rec, nil
}

func getRecords() ([]VpnRecord, error) {
	f, err := os.Open("list.csv")
	if err != nil {
		return []VpnRecord{}, err
	}

	defer f.Close()

	reader := bufio.NewReader(f)

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
			fmt.Println("error:", err)
			return []VpnRecord{}, err
		}

		record, err := makeRecordFromCsvRow(row)
		if err != nil {
			continue
		}
		vpnRecords = append(vpnRecords, record)
		fmt.Println(record)
	}

	return vpnRecords, nil
}

func saveRecord(vpn VpnRecord, savePath string) error {
	f, err := os.Create(savePath + vpn.Filename())
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(vpn.OpenVPNConfig)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	result, err := getRecords()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, rec := range result {
		fmt.Printf("saving '%s'\n", rec.Filename())
		err := saveRecord(rec, "./ovpn/")
		if err != nil {
			fmt.Println("ERROR:", err)
		}
	}
}
