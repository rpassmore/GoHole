package dnsserver

import (
    "fmt"
	"log"
	"net"
	"strings"
	"time"

    "github.com/miekg/dns"

    "GoHole/config"
    "GoHole/dnscache"
    "GoHole/logs"
    "GoHole/encryption"
)

func parseQuery(clientIp string, m *dns.Msg) {
	for _, q := range m.Question {
		var err error = nil
		var ip = ""
		cleanedName := q.Name[0:len(q.Name)-1] // remove the end "."
		qType := "A"
		isCached := false
		isBlocked := false
		isIpv4 := true

		if q.Qtype == dns.TypeA{
			ip, isBlocked, err = dnscache.GetDomainIPv4(cleanedName)
		}else if q.Qtype == dns.TypeAAAA{
			ip, isBlocked, err = dnscache.GetDomainIPv6(cleanedName)
			qType = "AAAA"
			isIpv4 = false
		}

		if ip != "" && err == nil {
			rr, err := dns.NewRR(fmt.Sprintf("%s %s %s", q.Name, qType, ip))
			if err == nil {
				m.Answer = append(m.Answer, rr)
			}
			isCached = true
		}else{
			// Request to a DNS server
			c := new(dns.Client)
			msg := new(dns.Msg)
			msg.SetQuestion(dns.Fqdn(q.Name), q.Qtype)
			msg.RecursionDesired = true

		    r, _, err := c.Exchange(msg, net.JoinHostPort(config.GetInstance().UpstreamDNSServer, "53"))
		    if r == nil {
		    	log.Printf("*** error: %s\n", err.Error())
		    	return
		    }

		    if r.Rcode != dns.RcodeSuccess {
		    	log.Printf(" *** invalid answer name %s after %s query for %s\n", q.Name, qType, q.Name)
		    	return
		    }
		    // Parse Answer
		    for _, a := range r.Answer {
		    	ans := strings.Split(a.String(), "\t")
		    	if len(ans) == 5 && ans[3] == qType{
		    		// Save on cache
		    		if q.Qtype == dns.TypeA{
		    			dnscache.AddDomainIPv4(cleanedName, ans[4], true)
					}else if q.Qtype == dns.TypeAAAA{
						dnscache.AddDomainIPv6(cleanedName, ans[4], true)
					}
		    	}
		    }
		    // Set answer for the client
		    m.Answer = r.Answer
		    isCached = false
		}

		// Add logs
		logs.AddQuery(clientIp, cleanedName, isCached, time.Now())
		go logs.AddQueryToGraphite(isBlocked, isIpv4, isCached)

		log.Printf("Query for %s from %s, blocked : %s, cached : %s", q.Name, clientIp, isBlocked, isCached)
	}
}

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false
	clientIp := w.RemoteAddr().String()
	clientIp = clientIp[0:strings.LastIndex(clientIp, ":")] // remove port

	switch r.Opcode {
	case dns.OpcodeQuery:
		parseQuery(clientIp, m)
	}

	w.WriteMsg(m)
}

func handleSecureDnsRequest(conn *net.UDPConn, buf []byte, addr net.UDPAddr){
	query, err := encryption.Decrypt(buf)
    if err != nil{
    	return
    }

    m := new(dns.Msg)
    m.Unpack(query)
    clientIp := addr.String()
    clientIp = clientIp[0:strings.LastIndex(clientIp, ":")] // remove port
    parseQuery(clientIp, m)

    reply, err := m.Pack()
    if err != nil{
    	return
    }
    eReply, err := encryption.Encrypt(reply)
    if err != nil{
    	return
    }

    conn.WriteToUDP(eReply, &addr)
}

func listenAndServeSecure(){
	serverAddr, err := net.ResolveUDPAddr("udp",":"+ config.GetInstance().SecureDNSPort)
	if err != nil {
		log.Fatalf("Failed to start DNS Secure Server: %s\n", err)
	}
	conn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		log.Fatalf("Failed to start DNS Secure Server: %s\n", err)
	}
	defer conn.Close()

	log.Printf("Starting Secure DNS Server at %s\n", config.GetInstance().SecureDNSPort)

	//simple read
	for{
		buf := make([]byte, 2048)
		n, addr, err := conn.ReadFromUDP(buf)
        if err != nil {
            continue
        }

        go handleSecureDnsRequest(conn, buf[:n], *addr)
	}
}

func ListenAndServe(){

	// add go.hole domain to our cache :)
	dnscache.AddDomainIPv4("go.hole", config.GetInstance().ServerIP, false)

	// start the graphite statistics loop
	go logs.StartStatsLoop()

	dns.HandleFunc(".", handleDnsRequest)
	// Start DNS server
	port := config.GetInstance().DNSPort


	ief, err := net.InterfaceByName("wlan0")
	if err !=nil{
		log.Fatal(err)
	}
	addrs, err := ief.Addrs()
	if err !=nil{
		log.Fatal(err)
	}

	fmt.Println("Using interface", addrs[0])
	server := &dns.Server{Addr: addrs[0].String() + ":" + port, Net: "udp"}

	//server := &dns.Server{Addr: ":" + port, Net: "udp"}


	log.Printf("Starting at %s\n", port)
	//go listenAndServeSecure()

	err = server.ListenAndServe()
	defer server.Shutdown()
	if err != nil {
		log.Fatalf("Failed to start DNS Server: %s\n ", err.Error())
	}

}