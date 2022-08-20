package main

import (
	"database/sql"
	"flag"
	_ "github.com/mattn/go-sqlite3"
	"hello/scraper/database"
	"hello/scraper/scrapers"
	"log"
	"net/http"
)

var (
	env *string
)

func init() {
	env = flag.String("output", "mibs.sqlite", "data source name")
}

func main() {
	flag.Parse()
	db, err := sql.Open("sqlite3", *env)
	if err != nil {
		log.Printf("could not connect to database: %v", err)
		return
	}

	sqlDb := database.NewSqlDb(db)
	err = sqlDb.Prepare()
	if err != nil {
		log.Printf("could not prepare database: %v", err)
		return
	}

	oid, err := sqlDb.GetLastOidCache()
	if err != nil {
		log.Printf("could not get last added oid: %v", err)
	}
	urlCache, err := sqlDb.FillCache()
	if err != nil {
		log.Printf("could not fill cache: %v", err)
	}

	parser := scrapers.NewOIDParser(urlCache, http.DefaultClient)
	scraper := scrapers.NewOIDScraper(oid, urlCache, sqlDb, parser)

	err = scraper.Start()
	if err != nil {
		log.Fatalf("scraper stopped: %v", err)
	}
	return
}
