package main

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func getDbPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	vpnlistDir := filepath.Join(configDir, "vpnlist")

	err = os.MkdirAll(vpnlistDir, 0o700)

	if err != nil {
		return "", err
	}

	return filepath.Join(vpnlistDir, "db.sqlite"), nil
}

func initDb() (*sql.DB, error) {
	dbFilePath, err := getDbPath()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbFilePath)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS servers (
id INTEGER PRIMARY KEY,
HostName       VARCHAR(255) UNIQUE,
IP             VARCHAR(255),
Score          INT,
Ping           INT,
Speed          INT,
CountryLong    VARCHAR(255),
CountryShort   VARCHAR(5),
NumVpnSessions INT,
Uptime         INT,
TotalUsers     INT,
TotalTraffic   INT,
LogType        VARCHAR(255),
Operator       VARCHAR(255),
Message        VARCHAR(255),
OpenVPNConfig  TEXT NOT NULL
);`); err != nil {
		return nil, err
	}

	return db, nil
}

func saveRecord(db *sql.DB, rec VpnRecord) error {
	_, err := db.Exec(`INSERT INTO servers(HostName, IP, Score, Ping, Speed, 
CountryLong, CountryShort, NumVpnSessions, Uptime, TotalUsers, TotalTraffic,
LogType, Operator, Message, OpenVPNConfig)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (HostName) DO UPDATE SET
IP=excluded.IP,
Score=excluded.Score,
Ping=excluded.Ping,
Speed=excluded.Speed,
CountryLong=excluded.CountryLong,
CountryShort=excluded.CountryShort,
NumVpnSessions=excluded.NumVpnSessions,
Uptime=excluded.Uptime,
TotalUsers=excluded.TotalUsers,
TotalTraffic=excluded.TotalTraffic,
Operator=excluded.Operator,
Message=excluded.Message,
OpenVPNConfig=excluded.OpenVPNConfig;`,
		rec.HostName, rec.IP, rec.Score, rec.Ping, rec.Speed, rec.CountryLong,
		rec.CountryShort, rec.NumVpnSessions, rec.Uptime, rec.TotalUsers,
		rec.TotalTraffic, rec.LogType, rec.Operator, rec.Message,
		rec.OpenVPNConfig)
	return err
}

func saveRecords(db *sql.DB, records []VpnRecord) error {
	for _, rec := range records {
		err := saveRecord(db, rec)
		if err != nil {
			return err
		}
	}

	return nil
}

func getTotalRecords(db *sql.DB) (int, error) {
	row := db.QueryRow("SELECT COUNT(*) FROM servers;")
	var total int
	err := row.Scan(&total)
	if err != nil {
		return total, err
	}
	return total, nil
}

func getRandomConfig(db *sql.DB) (string, error) {
	row := db.QueryRow(`SELECT OpenVPNConfig FROM servers ORDER BY RANDOM() LIMIT 1;`)
	var config string
	err := row.Scan(&config)
	if err != nil {
		return config, err
	}
	return config, nil
}

func getAllRecords(db *sql.DB) ([]VpnRecord, error) {
	rows, err := db.Query(`SELECT HostName, IP, Ping, Speed, CountryShort FROM servers
ORDER BY CountryShort, HostName;`)
	if err != nil {
		return []VpnRecord{}, err
	}

	defer rows.Close()

	result := []VpnRecord{}

	for rows.Next() {
		var r VpnRecord
		err := rows.Scan(&r.HostName, &r.IP, &r.Ping, &r.Speed, &r.CountryShort)
		if err != nil {
			return result, err
		}
		result = append(result, r)
	}

	err = rows.Err()
	if err != nil {
		return result, err
	}

	return result, nil
}
