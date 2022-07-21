package main

import (
	"fmt"
	"os"
)

// func saveRecord(vpn VpnRecord, savePath string) error {
// 	f, err := os.Create(savePath + vpn.Filename())
// 	if err != nil {
// 		return err
// 	}

// 	defer f.Close()

// 	_, err = f.Write(vpn.OpenVPNConfig)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

func main() {
	// result, err := downloadRecords()
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }

	// fmt.Printf("Got %d servers\n", len(result))

	// for _, rec := range result {
	// 	fmt.Printf("saving '%s'\n", rec.Filename())
	// 	err := saveRecord(rec, "./ovpn/")
	// 	if err != nil {
	// 		fmt.Println("ERROR:", err)
	// 	}
	// }

	db, err := initDb()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: can not initialize db (%s).\n", err)
		os.Exit(1)
	}

	records, err := downloadRecords(os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: can not download vpn list (%s).\n", err)
		os.Exit(1)
	}

	err = saveRecords(db, records)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: can not save records (%s).\n", err)
		os.Exit(1)
	}
}
