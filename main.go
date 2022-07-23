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

	rootCmd = makeRootCmd(db)

	updateCmd := makeUpdateCmd(db)
	randomCmd := makeRandomCmd(db)
	showCmd := makeShowCmd(db)
	countriesCmd := makeCountriesCmd(db)
	versionCmd := makeVersionCmd()

	rootCmd.AddCommand(updateCmd, randomCmd, showCmd, countriesCmd, versionCmd)
}

func main() {
	rootCmd.Execute()
}
