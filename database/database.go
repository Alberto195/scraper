package database

import (
	"database/sql"
	"fmt"
	"hello/scraper/models"
	"log"
	"sync"
)

type DB interface {
	Prepare(query string) (*sql.Stmt, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Ping() error
}

type SqlDb struct {
	db DB
	mu *sync.Mutex
}

func NewSqlDb(db DB) *SqlDb {
	return &SqlDb{db: db, mu: &sync.Mutex{}}
}

func (s *SqlDb) Prepare() error {
	err := s.testConnection()
	if err != nil {
		return fmt.Errorf("can`t test database: %v", err)
	}

	stmt, err := s.db.Prepare("CREATE TABLE IF NOT EXISTS mib " +
		"(uid INTEGER not null constraint mib_pk primary key autoincrement,oid VARCHAR(64) not null," +
		"name VARCHAR(64) not null,sub_ch INTEGER not null,sub_total INTEGER,descr VARCHAR(100)," +
		"inf VARCHAR(100) default '-')")
	if err != nil {
		return fmt.Errorf("cant prepare a query: %v", err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("cant execute a prepare query: %v", err)
	}

	stmt, err = s.db.Prepare("CREATE TABLE IF NOT EXISTS cacheUrls (id INTEGER not null constraint cacheUrls_pk primary key autoincrement, oid VARCHAR(20) not null)")
	if err != nil {
		return fmt.Errorf("cant prepare a query: %v", err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("cant execute a prepare query: %v", err)
	}

	return nil
}

func (s *SqlDb) Insert(oid string, info *models.TableInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	stmt, err := s.db.Prepare("INSERT INTO mib(oid, name, sub_ch, sub_total, descr, inf) values(?,?,?,?,?,?);")
	if err != nil {
		return fmt.Errorf("cant prepare a query: %v", err)
	}

	if info.Name != "" {
		_, err = stmt.Exec(oid, info.Name, info.SubCh, info.SubTotal, info.Desc, info.Inf)
		if err != nil {
			return fmt.Errorf("cant execute an insert query: %v", err)
		}
	}

	return nil
}

func (s *SqlDb) GetLastOidCache() (string, error) {
	var oid string

	s.mu.Lock()
	row := s.db.QueryRow("select oid from cacheUrls ORDER BY id DESC LIMIT 1;")
	err := row.Scan(&oid)
	if err != nil {
		s.mu.Unlock()
		return "/", fmt.Errorf("cant find last added oid: %v", err)
	}
	s.mu.Unlock()

	err = s.DeleteCache(oid)
	if err != nil {
		return "/", fmt.Errorf("cant delete last added oid: %v", err)
	}

	return oid, nil
}

func (s *SqlDb) InsertToCache(oid string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	stmt, err := s.db.Prepare("INSERT INTO cacheUrls(oid) values(?);")
	if err != nil {
		return fmt.Errorf("cant prepare a query: %v", err)
	}

	_, err = stmt.Exec(oid)
	if err != nil {
		return fmt.Errorf("cant execute an insert to cache query: %v", err)
	}

	return nil
}

func (s *SqlDb) FillCache() (*sync.Map, error) {
	urlCache := &sync.Map{}
	rows, err := s.db.Query("select oid from mib;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var oid string
		if err := rows.Scan(&oid); err != nil {
			log.Printf("Cache filled partly")
			return urlCache, err
		}
		urlCache.Store(oid, nil)
	}
	if err = rows.Err(); err != nil {
		return urlCache, err
	}

	log.Printf("Cache filled completely")
	return urlCache, nil
}

func (s *SqlDb) testConnection() error {
	if err := s.db.Ping(); err != nil {
		log.Fatalf("unable to reach database: %v", err)
		return err
	}

	return nil
}

func (s *SqlDb) DeleteCache(oid string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stmt, err := s.db.Prepare("DELETE FROM cacheUrls WHERE oid = (?);")
	if err != nil {
		return fmt.Errorf("cant prepare a query: %v", err)
	}

	_, err = stmt.Exec(oid)
	if err != nil {
		return fmt.Errorf("cant execute a delete cache query: %v", err)
	}

	return nil
}
