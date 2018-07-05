package dnscache

import (
	"time"

	"github.com/patrickmn/go-cache"
	"errors"
	"GoHole/config"
)

var instance *cache.Cache = nil

func GetInstance() *cache.Cache {
    if instance == nil {
		expireTime := time.Duration(config.GetInstance().DomainCacheTime)
		//purgeTime := time.Duration(config.GetInstance().Cache.PurgeTime)
    	instance = cache.New(expireTime*time.Minute, 10*time.Minute)
    	}

    return instance
}

func IPv4Preffix() string{
	return "ipv4:"
}
func IPv6Preffix() string{
	return "ipv6:"
}


func AddDomainIPv4(domain, ip string, expiration int) (error){
	if expiration > 0 {
		GetInstance().Set(IPv4Preffix() + domain, ip, cache.DefaultExpiration)
	} else {
		GetInstance().Set(IPv4Preffix() + domain, ip, cache.NoExpiration)
	}
	return nil
}

func AddDomainIPv6(domain, ip string, expiration int) (error) {
	if expiration > 0 {
		GetInstance().Set(IPv6Preffix()+domain, ip, cache.DefaultExpiration)
	} else {
		GetInstance().Set(IPv6Preffix()+domain, ip, cache.NoExpiration)
	}
	return nil
}

func DeleteDomainIPv4(domain string) (error){
	GetInstance().Delete(IPv4Preffix() + domain)
	return nil
}

func DeleteDomainIPv6(domain string) (error){
	GetInstance().Delete(IPv6Preffix() + domain)
	return nil
}

func GetDomainIPv4(domain string) (string, bool, error){
	ip, exp, found := GetInstance().GetWithExpiration(IPv4Preffix() + domain)
	if found == true {
		return ip.(string), exp.IsZero(), nil
	} else {
		return "", false, errors.New("domain " + domain + " not found")
	}

}


func GetDomainIPv6(domain string) (string, bool, error){
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

