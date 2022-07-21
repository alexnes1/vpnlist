package main

import (
	"fmt"
	"os"
)

func init() {
	db, err := initDb()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: can not initialize db (%s).\n", err)
		os.Exit(1)
	}

	updateCmd := makeUpdateCmd(db)

	rootCmd.AddCommand(updateCmd)
}

func main() {
	rootCmd.Execute()
}
