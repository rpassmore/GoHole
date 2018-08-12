/**
  Manage Black lists and white lists from a DB
*/

package domainLists

import (
	"github.com/asdine/storm"
	"log"
)

type DomainList struct {
	db *storm.DB
}

type ListEntry struct {
	Domain string `storm:"id,unique"`
	Allow  bool
}

func Open() *DomainList {
	var err error = nil
	dbPath := "./gohole-domains.db"

	log.Printf("Domain white/blacklist DB: %s", dbPath)

	db, err := storm.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	return &DomainList{db}
}

func (domainList *DomainList) Close() {
	if domainList.db != nil {
		domainList.Close()
	}
}

func (domainList *DomainList) WhiteListDomain(domain string) error {
	whiteListEntry := ListEntry{Domain: domain, Allow: true}
	return domainList.db.Save(&whiteListEntry)
}

func (domainList *DomainList) BlackListDomain(domain string) error {
	whiteListEntry := ListEntry{Domain: domain, Allow: false}
	return domainList.db.Save(&whiteListEntry)
}

func (domainList *DomainList) RemoveDomain(domain string) error {
	domainListEntry, err := domainList.FindDomain(domain)
	if err != nil {
		err = domainList.db.Drop(domainListEntry)
	}
	return err
}

func (domainList *DomainList) GetWhiteListedDomains() ([]ListEntry, error) {
	var domainEntries []ListEntry
	return domainEntries, domainList.db.Find("Allow", true, &domainEntries)
}

func (domainList *DomainList) GetBlackListedDomains() ([]ListEntry, error) {
	var domainEntries []ListEntry
	return domainEntries, domainList.db.Find("Allow", false, &domainEntries)
}

func (domainList *DomainList) FindDomain(domain string) (ListEntry, error) {
	var listEntry ListEntry
	return listEntry, domainList.db.One("Domain", domain, &listEntry)
}
