package sqldb

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type closeDB func() error

type DBContext interface {
	DB() *sqlx.DB
}

type dbContext struct {
	db *sqlx.DB
}

var _ DBContext = (*dbContext)(nil)

func NewDBContext(dsn string) (DBContext, closeDB, error) {
	// this Pings the database trying to connect
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, nil, err
	}
	return &dbContext{db: db},
		func() error {
			return db.Close()
		},
		nil
}

func (c *dbContext) DB() *sqlx.DB {
	return c.db
}
