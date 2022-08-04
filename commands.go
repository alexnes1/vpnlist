package main

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/go-ping/ping"
	"github.com/spf13/cobra"
)

var rootCmd *cobra.Command

func makeRootCmd(db *sql.DB) *cobra.Command {

	countries := []string{}
	var speed int
	var checkOnline bool
	var pingWorkers int
	var pingTimeout time.Duration

	cmd := &cobra.Command{
		Use:   "vpnlist",
		Short: "list all server records",
		Long: `This program is capable of parsing vpngate.net for OpenVPN configs, their storage and retrieval. 
Default command lists all servers stored in the database`,
		Run: func(cmd *cobra.Command, args []string) {
			records, err := getAllRecords(db, countries, speed)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: can not retrieve records (%s).\n", err)
				os.Exit(1)
			}

			if len(records) == 0 {
				fmt.Fprintln(os.Stdout, `There are no server records in the local database yet.
To populate the database, run 'vpnlist update'.`)
				return
			}

			if checkOnline {
				printRecordsWithPing(os.Stdout, records, pingWorkers, pingTimeout)
			} else {
				printRecords(os.Stdout, records)
			}
		},
	}

	cmd.Flags().StringSliceVarP(&countries, "country", "c", countries, "show records only with certain country code")
	cmd.Flags().IntVarP(&speed, "speed", "s", 0, "show records only with speed equal or greater (Mbps)")
	cmd.Flags().BoolVarP(&checkOnline, "ping", "p", false, "check if server is online")
	cmd.Flags().IntVarP(&pingWorkers, "ping-workers", "w", 1, "ping several servers simultaneously")
	cmd.Flags().DurationVarP(&pingTimeout, "ping-timeout", "t", 500*time.Millisecond, "ping timeout")

	return cmd
}

func printRecords(out io.Writer, records []VpnRecord) {
	fmt.Fprintf(out, "%-3s\t%-17s\t%-17s\t%-12s\n", "", "IP", "Host", "Speed")
	for _, r := range records {
		fmt.Fprintf(out, "%s\n", r)
	}
}

func isOnline(addr string, timeout time.Duration) (bool, time.Duration) {
	pinger, err := ping.NewPinger(addr)
	if err != nil {
		panic(err)
	}
	// pinger.SetPrivileged(true)
	pinger.Count = 1
	pinger.Timeout = timeout
	pinger.Run()
	stats := pinger.Statistics()
	return stats.PacketsSent == stats.PacketsRecv, stats.AvgRtt
}

const colorRed = "\033[31m"
const colorGreen = "\033[32m"
const colorReset = "\033[0m"

func pingRecords(from <-chan VpnRecord, to chan<- VpnRecord, wg *sync.WaitGroup, timeout time.Duration) {
	for record := range from {
		record.Online, record.AvgPing = isOnline(record.IP, timeout)
		to <- record
	}
	wg.Done()
}

func printPingedRecords(from <-chan VpnRecord, out io.Writer, wg *sync.WaitGroup) {
	for record := range from {
		if record.Online {
			fmt.Fprintf(out, "%s%s\tonline (%v)%s\n", colorGreen, record, record.AvgPing, colorReset)
		} else {
			fmt.Fprintf(out, "%s%s\toffline%s\n", colorRed, record, colorReset)
		}
	}
	wg.Done()
}

func printRecordsWithPing(out io.Writer, records []VpnRecord, pingWorkers int, pingTimeout time.Duration) {
	fmt.Fprintf(out, "%-3s\t%-17s\t%-17s\t%-12s\n", "", "IP", "Host", "Speed")

	const bufSize = 50
	toPing := make(chan VpnRecord, bufSize)
	pinged := make(chan VpnRecord, bufSize)

	wgPing := &sync.WaitGroup{}

	for i := 0; i < pingWorkers; i++ {
		wgPing.Add(1)
		go pingRecords(toPing, pinged, wgPing, pingTimeout)
	}

	wgPrint := &sync.WaitGroup{}
	wgPrint.Add(1)
	go printPingedRecords(pinged, out, wgPrint)

	for _, r := range records {
		toPing <- r
	}
	close(toPing)
	wgPing.Wait()
	close(pinged)
	wgPrint.Wait()
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
	countries := []string{}
	var speed int

	cmd := &cobra.Command{
		Use:   "random",
		Short: "get random config",
		Long:  "Get random OpenVPN config from the local database",
		Run: func(cmd *cobra.Command, args []string) {
			record, err := getRandomConfig(db, countries, speed)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: can not retrieve config (%s).\n", err)
				os.Exit(1)
			}

			fmt.Fprintf(os.Stdout, "# HOST: %s.opengw.net\n", record.HostName)
			fmt.Fprintf(os.Stdout, "# IP: %s\n", record.IP)
			fmt.Fprintf(os.Stdout, "# COUNTRY: %s\n", record.CountryLong)
			fmt.Fprintf(os.Stdout, "#%s\n", record.OpenVPNConfig)
		},
	}

	cmd.Flags().StringSliceVarP(&countries, "country", "c", countries, "show records only with certain country code")
	cmd.Flags().IntVarP(&speed, "speed", "s", 0, "show records only with speed equal or greater (Mbps)")

	return cmd
}

func makeShowCmd(db *sql.DB) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "show specific config",
		Long:  "show OpenVPN config for host with specified name",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			record, err := getSpecificConfig(db, args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: can not retrieve config (%s).\n", err)
				os.Exit(1)
			}

			fmt.Fprintf(os.Stdout, "# HOST: %s.opengw.net\n", record.HostName)
			fmt.Fprintf(os.Stdout, "# IP: %s\n", record.IP)
			fmt.Fprintf(os.Stdout, "# COUNTRY: %s\n", record.CountryLong)
			fmt.Fprintf(os.Stdout, "#%s\n", record.OpenVPNConfig)
		},
	}
}

func makeCountriesCmd(db *sql.DB) *cobra.Command {
	return &cobra.Command{
		Use:   "countries",
		Short: "list all countries",
		Long:  "List countries of records stored in the database",
		Run: func(cmd *cobra.Command, args []string) {
			countries, err := getCountries(db)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: can not retrieve countries (%s).\n", err)
				os.Exit(1)
			}

			for _, c := range countries {
				fmt.Fprintf(os.Stdout, "%s\n", c)
			}

		},
	}
}

func makeVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "version",
		Long:  "Print program version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(os.Stdout, "%s (built at %s with %s)\n", Version, BuildTime, runtime.Version())
		},
	}
}
