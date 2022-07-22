package main

import (
	"fmt"
	"os"
	"time"
)

var (
	Version   string
	BuildTime string = time.Now().Format(time.RFC3339)
)

func init() {
	db, err := initDb()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: can not initialize db (%s).\n", err)
		os.Exit(1)
	}

	updateCmd := makeUpdateCmd(db)
	randomCmd := makeRandomCmd(db)
	showCmd := makeShowCmd(db)
	allCmd := makeAllCmd(db)
	countriesCmd := makeCountriesCmd(db)

	rootCmd.AddCommand(updateCmd, randomCmd, showCmd, allCmd, countriesCmd)
}

func main() {
	rootCmd.Execute()
}
