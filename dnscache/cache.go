package dnscache

import (
	"time"

	"GoHole/config"
	"errors"
	"github.com/patrickmn/go-cache"
)

var instance *cache.Cache = nil

func GetInstance() *cache.Cache {
	if instance == nil {
		expireTime := time.Duration(config.GetInstance().DomainCacheTime)
		purgeTime := time.Duration(config.GetInstance().DomainPurgeInterval)
		instance = cache.New(expireTime*time.Second, purgeTime*time.Second)
	}

	return instance
}

func IPv4Preffix() string {
	return "ipv4:"
}
func IPv6Preffix() string {
	return "ipv6:"
}

func AddDomainIPv4(domain, ip string, expires bool) {
	if expires {
		GetInstance().Set(IPv4Preffix()+domain, ip, cache.DefaultExpiration)
	} else {
		GetInstance().Set(IPv4Preffix()+domain, ip, cache.NoExpiration)
	}
}

func AddDomainIPv6(domain, ip string, expires bool) {
	if expires {
		GetInstance().Set(IPv6Preffix()+domain, ip, cache.DefaultExpiration)
	} else {
		GetInstance().Set(IPv6Preffix()+domain, ip, cache.NoExpiration)
	}
}

func DeleteDomainIPv4(domain string) error {
	GetInstance().Delete(IPv4Preffix() + domain)
	return nil
}

func DeleteDomainIPv6(domain string) error {
	GetInstance().Delete(IPv6Preffix() + domain)
	return nil
}

func GetDomainIPv4(domain string) (string, bool, error) {
	ip, exp, found := GetInstance().GetWithExpiration(IPv4Preffix() + domain)
	if found == true {
		return ip.(string), exp.IsZero(), nil
	} else {
		return "", false, errors.New("domain " + domain + " not found")
	}

}

func GetDomainIPv6(domain string) (string, bool, error) {
	ip, exp, found := GetInstance().GetWithExpiration(IPv6Preffix() + domain)
	if found == true {
		return ip.(string), exp.IsZero(), nil
	} else {
		return "", false, errors.New("domain " + domain + " not found")
	}
}

func Flush() {
	GetInstance().Flush()
}
