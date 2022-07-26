package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	sq "github.com/Masterminds/squirrel"
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

func upperStringsSlice(a []string) []string {
	result := make([]string, 0, len(a))
	for _, s := range a {
		result = append(result, strings.ToUpper(s))
	}
	return result
}

func getRandomConfig(db *sql.DB, countries []string, speed int) (VpnRecord, error) {
	query := sq.Select("OpenVPNConfig", "HostName", "IP", "CountryLong").
		From("servers").OrderBy(`RANDOM()`).Limit(1)
	if len(countries) > 0 {
		countriesUpper := upperStringsSlice(countries)
		query = query.Where(sq.Eq{"CountryShort": countriesUpper})
	}

	if speed > 0 {
		query = query.Where(sq.Gt{"Speed": speed * 1000000})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return VpnRecord{}, err
	}

	row := db.QueryRow(sql, args...)
	var record VpnRecord
	err = row.Scan(&record.OpenVPNConfig, &record.HostName, &record.IP, &record.CountryLong)
	if err != nil {
		return VpnRecord{}, err
	}
	return record, nil
}

func getAllRecords(db *sql.DB, countries []string, speed int) ([]VpnRecord, error) {
	query := sq.Select("HostName", "IP", "Ping", "Speed", "CountryShort").
		From("servers").OrderBy("CountryShort", "HostName")

	if len(countries) > 0 {
		countriesUpper := upperStringsSlice(countries)
		query = query.Where(sq.Eq{"CountryShort": countriesUpper})
	}

	if speed > 0 {
		query = query.Where(sq.Gt{"Speed": speed * 1000000})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return []VpnRecord{}, err
	}

	rows, err := db.Query(sql, args...)
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

func getSpecificConfig(db *sql.DB, search string) (VpnRecord, error) {
	query := sq.Select("OpenVPNConfig", "HostName", "IP", "CountryLong").
		From("servers").Where("HostName LIKE ?", fmt.Sprint("%", search, "%")).Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return VpnRecord{}, err
	}

	row := db.QueryRow(sql, args...)

	var record VpnRecord
	err = row.Scan(&record.OpenVPNConfig, &record.HostName, &record.IP, &record.CountryLong)
	if err != nil {
		return VpnRecord{}, err
	}
	return record, nil
}

func getCountries(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`SELECT DISTINCT(CountryLong || ' (' || CountryShort || ')') AS country 
FROM servers ORDER BY country;`)
	if err != nil {
		return []string{}, err
	}

	defer rows.Close()

	result := []string{}

	for rows.Next() {
		var country string
		err := rows.Scan(&country)
		if err != nil {
			return result, err
		}
		result = append(result, country)
	}

	err = rows.Err()
	if err != nil {
		return result, err
	}

	return result, nil
}
