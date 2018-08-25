package main

import (
	"GoHole/domainLists"
	"GoHole/rest"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"

	"GoHole/config"
	"GoHole/dnscache"
	"GoHole/dnsserver"
	"GoHole/encryption"
	"GoHole/logs"
	"GoHole/parser"
)

/* Update version number on each release:
   Given a version number x.y.z, increment the:

   x - major release
   y - minor release
   z - build number
*/
const GOHOLE_VERSION = "1.1.0"

var (
	Commit          string
	CompilationDate string
	// Command line options
	port    = flag.String("p", "", "Set DNS server port")
	cfgFile = flag.String("c", "./config.json", "Config file")
	version = flag.Bool("v", false, "Show current GoHole version")

	// option to start the DNS server
	startDNS = flag.Bool("s", false, "Start DNS server")
	// option to stop the DNS server
	stopDNS = flag.Bool("stop", false, "Stop DNS server")

	// Add domain to blacklist by command line
	// example: gohole -ad google.com -ip4 0.0.0.0 -ip6 "::1"
	domainAdd = flag.String("ad", "", "Domain to add")
	ipv4      = flag.String("ip4", "", "IPv4 Address for the domain")
	ipv6      = flag.String("ip6", "", "IPv6 Address for the domain")

	// Delete domain from blacklist/cache by command line
	// example: gohole -dd google.com
	domainDelete = flag.String("dd", "", "Domain to delete")

	// Flush Cache&Blacklist DB (RedisDB)
	// example: gohole -fcache
	flushCache = flag.Bool("fcache", false, "Flush domain cache")

	// Parse blacklist of domains and add to the cache server
	// example: gohole -ab http://domain/path/to/list.txt
	// example: gohole -ab /path/to/list.txt
	blacklistFile = flag.String("ab", "", "Path to blacklist file")

	// Parse blacklist's list and add to the cache server
	// example: gohole -abl /path/to/list_of_blacklists.txt
	blacklistslistFile = flag.String("abl", "", "Path to list of blacklists file (one list per line)")

	// Show queries by client IP
	// example: gohole -lip 127.0.0.1
	listip = flag.String("lip", "", "Show queries by client IP")

	// Show queries by domain
	// example: gohole -ld 127.0.0.1
	listdomain = flag.String("ld", "", "Show queries by domain")

	// Show top domain queries
	// example: gohole -ldt -limit 10
	listdomaintop = flag.Bool("ldt", false, "Show top domain queries")

	// Show clients
	// example: gohole -lc
	listclients = flag.Bool("lc", false, "Show clients")
	listLimit   = flag.Int("limit", 100, "Number of registers to show for arguments: -lip")

	// Flush queries log
	// example: gohole -flog
	flushLog = flag.Bool("flog", false, "Flush queries log")

	// Generate Encryption Key
	// example: gohole -gkey
	gkey = flag.Bool("gkey", false, "Generate encryption key and export in 'enc.key' file")
)

func showVersionInfo() {
	fmt.Println("----------------------------------------")
	fmt.Printf("GoHole v%s\nCommit: %s\nCompilation date: %s\n", GOHOLE_VERSION, Commit, CompilationDate)
	fmt.Println("----------------------------------------")
}

func main() {

	//start log DB
	logs := logs.Open()
	//defer logs.Close()

	//Start blacklist and white list DB
	domainLists := domainLists.Open()
	//defer domainLists.Close()

	//Start rest service
	go rest.NewRestService(logs, domainLists)

	dnsServer := dnsserver.NewDnsServer(logs)

	flag.Parse()

	config.CreateInstance(*cfgFile)
	if *port != "" {
		config.GetInstance().DNSPort = *port
	}

	encryption.CreateInstance()
	if *gkey {
		k, err := encryption.GenerateRandomKey()
		if err != nil {
			log.Printf("Error generating key: %s", err)
			return
		}
		encryption.ExportKeyToFile(k, "enc.key")
	}
	encryption.ImportKeyFromFile(config.GetInstance().EncryptionKey)

	if *version {
		showVersionInfo()
	}

	if *domainAdd != "" && *ipv4 != "" && *ipv6 != "" {
		dnscache.AddDomainIPv4(*domainAdd, *ipv4, false)
		dnscache.AddDomainIPv6(*domainAdd, *ipv6, false)
	}
	if *domainDelete != "" {
		err := dnscache.DeleteDomainIPv4(*domainDelete)
		if err != nil {
			log.Printf("Error: %s", err)
		}
		err = dnscache.DeleteDomainIPv6(*domainDelete)
		if err != nil {
			log.Printf("Error: %s", err)
		}
	}
	if *flushCache {
		dnscache.Flush()
		log.Printf("Cache flushed!")
	}

	if *blacklistFile != "" {
		parser.ParseBlacklistFile(*blacklistFile)
	}
	if *blacklistslistFile != "" {
		parser.ParseBlacklistsListFile(*blacklistslistFile)
	}

	if *listip != "" {
		queries, err := logs.GetQueriesByClientIp(*listip, *listLimit)
		if err != nil {
			log.Printf("Error: %s", err)
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Client IP", "Domain", "Date"})
			for _, q := range queries {
				toTime := q.Timestamp.Format(time.RFC1123)
				table.Append([]string{q.ClientIp, q.Domain, toTime})
			}
			table.Render()
		}
	}
	if *listdomain != "" {
		queries, err := logs.GetQueriesByDomain(*listdomain)
		if err != nil {
			log.Printf("Error: %s", err)
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Client IP", "Domain", "Date"})
			for _, q := range queries {
				toTime := q.Timestamp.Format(time.RFC1123)
				table.Append([]string{q.ClientIp, q.Domain, toTime})
			}
			table.Render()
		}
	}
	if *listdomaintop {
		domains, err := logs.GetTopDomains(*listLimit)
		if err != nil {
			log.Printf("Error: %s", err)
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Domain", "Num. Queries"})
			for _, c := range domains {
				table.Append([]string{c.Domain, strconv.Itoa(c.Queries)})
			}
			table.Render()
		}
	}
	if *listclients {
		clients, err := logs.GetTopClients()
		if err != nil {
			log.Printf("Error: %s", err)
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Client IP", "Num. Queries"})
			for _, c := range clients {
				table.Append([]string{c.ClientIp, strconv.Itoa(c.Queries)})
			}
			table.Render()
		}
	}
	if *flushLog {
		err := logs.Flush()
		if err != nil {
			log.Printf("Error: %s", err)
		} else {
			log.Printf("Query logs flushed!")
		}
	}

	if *startDNS {
		dnsServer.ListenAndServe()
	}
	if *stopDNS {
		exec.Command("killall", os.Args[0]).Run()
	}

}
