package database

import (
	"database/sql"
	"github.com/DavidHuie/gomigrate"
	"github.com/go-sql-driver/mysql"
	"ims-release/config"
)

// a binary storage provider is used to store and
// retrieve binary data under a unique key
type DB interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type DbHandle struct {
	inner *sql.DB
}

var ErrNoRows = sql.ErrNoRows

func generateDbConnString(c *config.Config) string {
	config := mysql.Config{
		User:      c.DbUser,
		Passwd:    c.DbPassword,
		DBName:    c.DbName,
		Net:       c.DbProtocol,
		Addr:      c.DbAddress,
		ParseTime: true,
	}
	return config.FormatDSN()
}

func NewDbHandle(c *config.Config) (DbHandle, error) {
	innerDb, err := sql.Open("mysql", generateDbConnString(c))
	db := DbHandle{innerDb}
	return db, err
}

func (db DbHandle) Migrate(migrationsPath string) error {
	migrator, _ := gomigrate.NewMigrator(db.inner, gomigrate.Mysql{}, migrationsPath)
	return migrator.Migrate()
}

func (db DbHandle) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.inner.QueryRow(query, args...)
}

func (db DbHandle) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.inner.Query(query, args...)
}

func (db DbHandle) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.inner.Exec(query, args...)
}
