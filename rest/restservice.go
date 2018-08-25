package rest

import (
	"GoHole/domainLists"
	"GoHole/logs"
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"log"
	"net/http"

	"time"
)

type RestService struct {
	logs.DBLogs
	*domainLists.DomainList
}

func NewRestService(dbLogs logs.DBLogs, dLists *domainLists.DomainList) *RestService {
	restService := RestService{dbLogs, dLists}

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Get("/latestQuery", restService.GetLatestQueryHandler)
	router.Get("/topClients", restService.GetTopClientsHandler)
	router.Get("/topDomains", restService.GetTopDomainsHandler)
	router.Get("/topBlockedDomains", restService.GetTopBlockedDomainsHandler)
	router.Get("/countQueriesBlocked", restService.GetCountQueriesBlockedHandler)
	router.Get("/queriesLast24hrs", restService.GetQueriesLast24hrsHandler)
	router.Get("/countQueries", restService.GetCountQueriesHandler)

	router.Get("/queriesForClient/{clientIp}", restService.GetQueriesByClientIpHandler)
	router.Get("/queriesByDomain/{domain}", restService.GetQueriesByDomainHandler)
	log.Fatal(http.ListenAndServe(":8080", router))
	return &restService
}

func (rest *RestService) getLimit(r *http.Request) int {
	limit, ok := r.Context().Value("limit").(int)
	if !ok {
		limit = -1
	}
	return limit
}

func (rest *RestService) GetCountQueriesBlockedHandler(w http.ResponseWriter, r *http.Request) {
	total, blocked, cached, err := rest.DBLogs.CountQueriesBlocked()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}
	type Totals struct {
		Total   int `json:"total"`
		Blocked int `json:"blocked"`
		Cached  int `json:"cached"`
	}
	totals := Totals{total, blocked, cached}
	json.NewEncoder(w).Encode(totals)
}

func (rest *RestService) GetQueriesLast24hrsHandler(w http.ResponseWriter, r *http.Request) {
	queryLogs, err := rest.DBLogs.GetQueriesSince(time.Now().Add(-24 * time.Hour))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}
	json.NewEncoder(w).Encode(queryLogs)
}

func (rest *RestService) GetCountQueriesHandler(w http.ResponseWriter, r *http.Request) {
	count, err := rest.DBLogs.CountQueries()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}
	json.NewEncoder(w).Encode(&count)
}

func (rest *RestService) GetQueriesByDomainHandler(w http.ResponseWriter, r *http.Request) {
	queryLogs, err := rest.DBLogs.GetQueriesByDomain(chi.URLParam(r, "domain"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}
	json.NewEncoder(w).Encode(queryLogs)
}

func (rest *RestService) GetQueriesByClientIpHandler(w http.ResponseWriter, r *http.Request) {
	queryLogs, err := rest.DBLogs.GetQueriesByClientIp(chi.URLParam(r, "clientIp"), rest.getLimit(r))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}
	json.NewEncoder(w).Encode(queryLogs)
}

func (rest *RestService) GetLatestQueryHandler(w http.ResponseWriter, r *http.Request) {
	queryLogs, err := rest.DBLogs.GetLatestQuery()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}
	json.NewEncoder(w).Encode(queryLogs)
}

func (rest *RestService) GetTopClientsHandler(w http.ResponseWriter, r *http.Request) {
	queryLogs, err := rest.DBLogs.GetTopClients()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}
	json.NewEncoder(w).Encode(queryLogs)
}

func (rest *RestService) GetTopDomainsHandler(w http.ResponseWriter, r *http.Request) {
	queryLogs, err := rest.DBLogs.GetTopDomains(rest.getLimit(r))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}
	json.NewEncoder(w).Encode(queryLogs)
}

func (rest *RestService) GetTopBlockedDomainsHandler(w http.ResponseWriter, r *http.Request) {
	queryLogs, err := rest.DBLogs.GetTopDomainsBlocked(rest.getLimit(r), true)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}
	json.NewEncoder(w).Encode(queryLogs)
}
