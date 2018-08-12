package rest

import (

"GoHole/domainLists"
"GoHole/logs"
"encoding/json"
"github.com/gorilla/mux"
"log"
"net/http"
"strconv"
"time"

)

type RestService struct {
  logs.DBLogs
  *domainLists.DomainList
}

func NewRestService(dbLogs logs.DBLogs, dLists *domainLists.DomainList) *RestService {
  restService := RestService{dbLogs, dLists}

  router := mux.NewRouter()
  router.HandleFunc("/latest_query", restService.GetLatestQuery).Methods("GET")
  router.HandleFunc("/top_clients", restService.GetTopClients).Methods("GET")
  router.HandleFunc("/top_domains", restService.GetTopDomains).Queries("limit", "{limit}").Methods("GET")
  router.HandleFunc("/top_blocked_domains", restService.GetTopBlockedDomains).Queries("blocked", "{blocked}", "limit", "{limit}").Methods("GET")
  router.HandleFunc("/count_queries_blocked", restService.GetCountQueriesBlocked).Methods("GET")
  router.HandleFunc("/queries_last_24hrs", restService.GetQueriesLast24hrs).Methods("GET")
  router.HandleFunc("/count_queries", restService.GetCountQueries).Methods("GET")

  router.HandleFunc("/queries_for_client/{clientIp}", restService.GetQueriesByClientIp).Queries("limit", "{limit}").Methods("GET")
  router.HandleFunc("/queries_by_domain/{domain}", restService.GetQueriesByDomain).Methods("GET")
  log.Fatal(http.ListenAndServe(":8080", router))
  return &restService
}

func (rest *RestService) getLimit(r *http.Request) int {
  limit, err := strconv.Atoi(r.FormValue("limit"))
  if err != nil {
    limit = -1
  }
  return limit
}

func (rest *RestService) GetCountQueriesBlocked(w http.ResponseWriter, r *http.Request) {
  total, blocked, cached, err := rest.DBLogs.CountQueriesBlocked()
  if err != nil {
   w.WriteHeader(http.StatusBadRequest)
   w.Write([]byte(err.Error()))
  }
  type Totals struct {
   Total int `json:"total"`
   Blocked int `json:"blocked"`
   Cached int `json:"cached"`
  }
  totals := Totals{total, blocked, cached}
  json.NewEncoder(w).Encode(totals)
}

func (rest *RestService) GetQueriesLast24hrs(w http.ResponseWriter, r *http.Request) {
  queryLogs, err := rest.DBLogs.GetQueriesSince(time.Now().Add(-24*time.Hour))
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(err.Error()))
  }
  json.NewEncoder(w).Encode(queryLogs)
}

func (rest *RestService) GetCountQueries(w http.ResponseWriter, r *http.Request) {
  count, err := rest.DBLogs.CountQueries()
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(err.Error()))
  }
  json.NewEncoder(w).Encode(&count)
}

func (rest *RestService) GetQueriesByDomain(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)

  queryLogs, err := rest.DBLogs.GetQueriesByDomain(vars["domain"])
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(err.Error()))
  }
  json.NewEncoder(w).Encode(queryLogs)
}

func (rest *RestService) GetQueriesByClientIp(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)

  queryLogs, err := rest.DBLogs.GetQueriesByClientIp(vars["clientIp"], rest.getLimit(r))
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(err.Error()))
  }
  json.NewEncoder(w).Encode(queryLogs)
}

func (rest *RestService) GetLatestQuery(w http.ResponseWriter, r *http.Request) {
  queryLogs, err := rest.DBLogs.GetLatestQuery()
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(err.Error()))
  }
  json.NewEncoder(w).Encode(queryLogs)
}

func (rest *RestService) GetTopClients(w http.ResponseWriter, r *http.Request) {
  queryLogs, err := rest.DBLogs.GetTopClients()
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(err.Error()))
  }
  json.NewEncoder(w).Encode(queryLogs)
}

func (rest *RestService) GetTopDomains(w http.ResponseWriter, r *http.Request) {
  queryLogs, err := rest.DBLogs.GetTopDomains(rest.getLimit(r))
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(err.Error()))
  }
  json.NewEncoder(w).Encode(queryLogs)
}

func (rest *RestService) GetTopBlockedDomains(w http.ResponseWriter, r *http.Request) {
  queryLogs, err := rest.DBLogs.GetTopDomainsBlocked(rest.getLimit(r), true)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(err.Error()))
  }
  json.NewEncoder(w).Encode(queryLogs)
}

