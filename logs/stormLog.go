package logs

import (
  "github.com/asdine/storm/q"
  "os/user"
  "log"

  "github.com/asdine/storm"
  "time"
)

type DBLogs interface {
  AddQuery(clientIp string, domain string, cached bool, blocked bool, timestamp time.Time) (error)
  AddClientIp(clientIp string, blocked bool) (error)
  AddDomain(domain string, blocked bool) (error)
  GetLatestQuery() ([]QueryLog, error)
  GetQueriesSince(from time.Time) ([]QueryLog, error)
  GetQueriesByClientIp(clientIp string, limit int) ([]QueryLog, error)
  GetQueriesByDomain(domain string) ([]QueryLog, error)
  CountQueries() (int, error)
  CountQueriesBlocked() (int, int, int, error)
  GetTopClients() ([]ClientLog, error)
  GetTopDomains(limit int) ([]DomainLog, error)
  GetTopDomainsBlocked(limit int, blocked bool) ([]DomainLog, error)
  Flush() (error)
  Close()
}

type dbLogsImpl struct {
  db *storm.DB
}

type QueryLog struct {
  Id        int `storm:"id,increment"`
  ClientIp  string `storm:"index"`
  Domain    string `storm:"index"`
  Cached    bool
  Blocked   bool
  Timestamp time.Time `storm:"index"`
}

type ClientLog struct {
  ClientIp string `storm:"id,unique"`
  Queries  int    `storm:"index"`
  Blocks   int
}

type DomainLog struct {
  Domain  string `storm:"id,unique"`
  Queries int
  Blocked bool
}

func Open() DBLogs {
  var err error = nil
  var dbPath string = ""

  usr, err := user.Current()
  if err != nil {
    dbPath = "./gohole-logs.db"
  } else {
    dbPath = usr.HomeDir + "/gohole-logs.db"
  }

  log.Printf("Logs DB: %s", dbPath)

  db, err := storm.Open(dbPath)
  if err != nil {
    log.Fatal(err)
  }
  return &dbLogsImpl{db}
}

func (dbLogs *dbLogsImpl) AddQuery(clientIp string, domain string, cached bool, blocked bool, timestamp time.Time) (error) {
  queryLog := QueryLog{ClientIp: clientIp, Domain: domain, Cached: cached, Blocked: blocked, Timestamp: timestamp}
  err := dbLogs.db.Save(&queryLog)
  if err != nil {
    return err
  }

  err = dbLogs.AddClientIp(clientIp, blocked)
  if err != nil {
    return err
  }

  err = dbLogs.AddDomain(domain, blocked)
  return err
}

func (dbLogs *dbLogsImpl) AddClientIp(clientIp string, blocked bool) (error) {
  var clientLog ClientLog
  err := dbLogs.db.One("ClientIp", clientIp, &clientLog)
  if err == nil {
    clientLog.ClientIp = clientIp
    clientLog.Queries = 0
  }

  if blocked {
    clientLog.Blocks = clientLog.Blocks + 1
  }

  clientLog.Queries = clientLog.Queries + 1
  err = dbLogs.db.Save(&clientLog)
  return err
}

func (dbLogs *dbLogsImpl) AddDomain(domain string, blocked bool) (error) {
  var domainLog DomainLog
  err := dbLogs.db.One("Domain", domain, &domainLog)
  if err != nil {
    domainLog.Domain = domain
    domainLog.Queries = 0
  }

  domainLog.Queries = domainLog.Queries + 1
  domainLog.Blocked = blocked
  err = dbLogs.db.Save(&domainLog)
  return err
}

func (dbLogs *dbLogsImpl) GetQueries() (int, error) {
  var queryLogs [] QueryLog
  count, err := dbLogs.db.Count(&queryLogs)
  return count, err
}

func (dbLogs *dbLogsImpl) CountQueries() (int, error) {
  var queryLogs [] QueryLog
  err := dbLogs.db.All(&queryLogs)
  return len(queryLogs), err
}

func (dbLogs *dbLogsImpl) CountQueriesBlocked() (int, int, int, error) {
  var queryLogs [] QueryLog
  total, err := dbLogs.CountQueries()
  if err != nil {
    return 0, 0, 0, err
  }
  err = dbLogs.db.Find("Blocked", true, &queryLogs)
  if err != nil {
    return 0, 0, 0, err
  }
  blockedCount := len(queryLogs)
  err = dbLogs.db.Find("Cached", true, &queryLogs)
  if err != nil {
    return 0, 0, 0, err
  }
  cachedCount := len(queryLogs)
  return total, blockedCount, cachedCount, err
}

func (dbLogs *dbLogsImpl) GetLatestQuery() ([]QueryLog, error) {
  var queryLogs [] QueryLog
  err := dbLogs.db.All(&queryLogs, storm.Limit(1), storm.Reverse())
  return queryLogs, err
}

func (dbLogs *dbLogsImpl) GetQueriesSince(from time.Time) ([]QueryLog, error) {
  var queryLogs [] QueryLog
  err := dbLogs.db.Select( q.Gte("Timestamp", from) ).OrderBy("Timestamp").Find(&queryLogs)
  return queryLogs, err
}

func (dbLogs *dbLogsImpl) GetQueriesByClientIp(clientIp string, limit int) ([]QueryLog, error) {
  var queryLogs [] QueryLog
  err := dbLogs.db.Find("ClientIp", clientIp, &queryLogs, storm.Limit(limit), storm.Reverse())
  return queryLogs, err
}

func (dbLogs *dbLogsImpl) GetQueriesByDomain(domain string) ([]QueryLog, error) {
  var queryLogs [] QueryLog
  err := dbLogs.db.Find("Domain", domain, &queryLogs, storm.Reverse())
  return queryLogs, err
}

func (dbLogs *dbLogsImpl) GetTopClients() ([]ClientLog, error) {
  var clientLogs [] ClientLog
  err := dbLogs.db.Select(q.Gte("Queries",0)).OrderBy("Queries").Reverse().Find(&clientLogs)
  return clientLogs, err
}



func (dbLogs *dbLogsImpl) GetTopDomains(limit int) ([]DomainLog, error) {
  var domainLogs [] DomainLog
  err := dbLogs.db.Select(q.Gte("Queries", 0)).OrderBy("Queries").Reverse().Limit(limit).Find(&domainLogs)
  return domainLogs, err
}

func (dbLogs *dbLogsImpl) GetTopDomainsBlocked(limit int, blocked bool) ([]DomainLog, error) {
  var domainLogs [] DomainLog
  err := dbLogs.db.Find("Blocked", blocked, &domainLogs, storm.Reverse(), storm.Limit(limit))
  return domainLogs, err
}

func (dbLogs *dbLogsImpl) Flush() (error) {
  var queryLog QueryLog
  var clientLog ClientLog
  var domainLog DomainLog
  dbLogs.db.Drop(queryLog)
  dbLogs.db.Drop(clientLog)
  dbLogs.db.Drop(domainLog)
  return nil
}

func (dbLogs *dbLogsImpl) Close() {
  if dbLogs != nil {
    dbLogs.db.Close()
  }
}

