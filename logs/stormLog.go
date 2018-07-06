package logs

import (
  "os/user"
  "log"

  "github.com/asdine/storm"
  "time"
)

type QueryLog struct {
  Id        int `storm:"id,increment"`
  ClientIp  string `storm:"index"`
  Domain    string `storm:"index"`
  Cached    bool
  Timestamp time.Time `storm:"index"`
}

type ClientLog struct {
  ClientIp string `storm:"id,unique"`
  Queries  int
}

type DomainLog struct {
  Domain  string `storm:"id,unique"`
  Queries int
}

var instance *storm.DB = nil

func GetInstance() *storm.DB {
  if instance == nil {
    var err error = nil
    var dbPath string = ""

    usr, err := user.Current()
    if err != nil {
      dbPath = "./gohole.db"
    } else {
      dbPath = usr.HomeDir + "/gohole.db"
    }

    log.Printf("Logs DB: %s", dbPath)

    instance, err = storm.Open(dbPath)
    if err != nil {
      log.Fatal(err)
    }
    //defer instance.Close()
  }

  return instance
}

func AddQuery(clientIp string, domain string, cached bool, timestamp time.Time) (error) {
  queryLog := QueryLog{ClientIp: clientIp, Domain: domain, Cached: cached, Timestamp: timestamp}
  err := GetInstance().Save(&queryLog)
  if err != nil {
    return err
  }

  err = AddClientIp(clientIp)
  if err != nil {
    return err
  }

  err = AddDomain(domain)
  return err
}

func AddClientIp(clientip string) (error) {
  var clientLog ClientLog
  err := GetInstance().One("ClientIp", clientip, &clientLog)
  if err != nil {
    clientLog.ClientIp = clientip
    clientLog.Queries = 0
  }

  clientLog.Queries = clientLog.Queries + 1
  err = GetInstance().Save(&clientLog)
  return err
}

func AddDomain(domain string) (error) {
  var domainLog DomainLog
  err := GetInstance().One("Domain", domain, &domainLog)
  if err != nil {
    domainLog.Domain = domain
    domainLog.Queries = 0
  }

  domainLog.Queries = domainLog.Queries + 1
  err = GetInstance().Save(&domainLog)
  return err
}

func GetLatestQuery() ([]QueryLog, error) {
  var queryLogs [] QueryLog
  err := GetInstance().All(&queryLogs, storm.Limit(10), storm.Reverse())
  return queryLogs, err
}

func GetQueriesByClientIp(clientIp string, limit int) ([]QueryLog, error) {
  var queryLogs [] QueryLog
  err := GetInstance().Find("ClientIp", clientIp, &queryLogs, storm.Limit(limit), storm.Reverse())
  return queryLogs, err
}

func GetQueriesByDomain(domain string) ([]QueryLog, error) {
  var queryLogs [] QueryLog
  err := GetInstance().Find("Domain", domain, &queryLogs, storm.Reverse())
  return queryLogs, err
}

func GetClients() ([]ClientLog, error) {
  var clientLogs [] ClientLog
  err := GetInstance().Select().OrderBy("Queries").Reverse().Find(&clientLogs)
  return clientLogs, err
}

func GetTopDomains(limit int) ([]DomainLog, error) {
  var domainLogs [] DomainLog
  err := GetInstance().Select().OrderBy("Queries").Reverse().Limit(limit).Find(&domainLogs)
  return domainLogs, err
}

func Flush() (error) {
  var queryLog QueryLog
  var clientLog ClientLog
  var domainLog DomainLog
  GetInstance().Drop(queryLog)
  GetInstance().Drop(clientLog)
  GetInstance().Drop(domainLog)
  return nil
}

