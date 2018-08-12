package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

// MyConfig struct
// This is the struct that the config.json must have
type MyConfig struct {
	Interface     string // the external interface to listen on
	ServerIP      string // the DNS server IP to redirect blocked ads
	DNSPort       string // listen on port
	SecureDNSPort string // listen port for encrypted DNS Server
	EncryptionKey string // Path to the encryption key file

	// Graphite info
	Graphite GraphiteConfig

	UpstreamDNSServer   string
	DomainCacheTime     int // time to save domains in cache (in seconds)
	DomainPurgeInterval int // interval at which expired domains are purged
}

// Graphite Config
type GraphiteConfig struct {
	Host string
	Port int
}

var instance *MyConfig = nil

func CreateInstance(filename string) *MyConfig {
	var err error
	instance, err = loadConfig(filename)
	if err != nil {
		log.Printf("Error loading config file: %s\nUsing default config.", err)
		// use defaults
		instance = &MyConfig{
			Interface:           "wlan0",
			ServerIP:            "0.0.0.0",
			DNSPort:             "53",
			SecureDNSPort:       "443",
			EncryptionKey:       "enc.key",
			UpstreamDNSServer:   "8.8.8.8",
			DomainCacheTime:     1800,
			DomainPurgeInterval: 600,
			Graphite: GraphiteConfig{
				Host: "localhost",
				Port: 2003,
			},
		}
	}

	return instance
}

func GetInstance() *MyConfig {
	return instance
}

func loadConfig(filename string) (*MyConfig, error) {
	var s *MyConfig

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return s, err
	}
	// Unmarshal json
	err = json.Unmarshal(bytes, &s)
	return s, err
}
