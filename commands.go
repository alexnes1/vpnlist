package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "vpnlist",
	Short: "OpenVPN configs grabber and storage for vpngate.net",
	Long:  "This program is capable of parsing vpngate.net for OpenVPN configs, their storage and retrieval",
}

func makeUpdateCmd(db *sql.DB) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "get vpn servers info from vpngate.net",
		Long:  "Get vpn servers info from vpngate.net and save it locally",
		Run: func(cmd *cobra.Command, args []string) {
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

			total, err := getTotalRecords(db)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: can not count records (%s).\n", err)
				os.Exit(1)
			}

			fmt.Fprintf(os.Stdout, "Got servers: %d, total servers in the database: %d.\n",
				len(records), total)
		},
	}
}

func makeRandomCmd(db *sql.DB) *cobra.Command {
	return &cobra.Command{
		Use:   "random",
		Short: "get random config",
		Long:  "Get random OpenVPN config from the local database",
		Run: func(cmd *cobra.Command, args []string) {
			config, err := getRandomConfig(db)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: can retrieve config (%s).\n", err)
				os.Exit(1)
			}

			fmt.Fprintf(os.Stdout, "%s\n", config)
		},
	}
}
