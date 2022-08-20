package scrapers

import (
	"hello/scraper/models"
	"log"
	"sync"
)

const (
	baseUrl      = "https://oidref.com"
	numDigesters = 5
	numWalkers   = 5
)

type Parser interface {
	Parse(chan string, chan<- string, chan<- string) error
}

type SqlDb interface {
	Insert(string, *models.TableInfo) error
	DeleteCache(string) error
	InsertToCache(string) error
	GetLastOidCache() (string, error)
}

type OIDScraper struct {
	startUrl string
	urlCache *sync.Map
	db       SqlDb
	parser   Parser
}

func NewOIDScraper(startUrl string, urlCache *sync.Map, db SqlDb, parser Parser) *OIDScraper {
	return &OIDScraper{
		startUrl: startUrl,
		urlCache: urlCache,
		db:       db,
		parser:   parser,
	}
}

func (s *OIDScraper) Start() error {
	paths, pathsToCache, errCh := s.walk()
	defer close(paths)
	defer close(pathsToCache)
	defer close(errCh)

	for i := 0; i < numDigesters; i++ {
		go func() {
			s.digesterCache(pathsToCache, errCh)
		}()
		go func() {
			s.digester(paths, errCh)
		}()
	}

	select {
	case err := <-errCh:
		return err
	}
}

func (s *OIDScraper) walk() (chan string, chan string, chan error) {
	paths := make(chan string)
	pathsToCache := make(chan string)
	errCh := make(chan error, 1)
	urlCh := make(chan string, numWalkers)

	for i := 0; i < numWalkers; i++ {
		go func(urlCh chan string) {
			err := s.parser.Parse(urlCh, paths, pathsToCache)
			if err != nil {
				close(urlCh)
				errCh <- err
			}
		}(urlCh)
	}
	go func(urlCh chan string) {
		for {
			restartUrl, err := s.db.GetLastOidCache()
			if err != nil && err.Error() != "cant find last added oid: sql: no rows in result set" {
				errCh <- err
				close(urlCh)
			}
			urlCh <- restartUrl
		}
	}(urlCh)

	return paths, pathsToCache, errCh
}

func (s *OIDScraper) digester(paths <-chan string, errCh chan<- error) {
	for path := range paths {
		err := s.db.DeleteCache(path)
		if err != nil {
			errCh <- err
		}
		if v, ok := s.urlCache.Load(path); ok {
			err = s.db.Insert(path, v.(*models.TableInfo))
			if err != nil {
				errCh <- err
			}
			log.Println("Added to database info from link ", path)
		} else {
			log.Println("No info found for url ", path)
		}
	}
}

func (s *OIDScraper) digesterCache(paths <-chan string, errCh chan<- error) {
	for path := range paths {
		err := s.db.InsertToCache(path)
		if err != nil {
			errCh <- err
		}
	}
}
