package connection

import (
	"auth/internal/config"
	"database/sql"
	"fmt"
	"log"

	"github.com/doug-martin/goqu"
	_ "github.com/lib/pq"
	_ "gopkg.in/doug-martin/goqu.v5/adapters/postgres"
)

func GetDatabase(conf config.Database) (*goqu.Database, *sql.DB) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=%s",
		conf.HOST, conf.PORT, conf.USER, conf.PASS, conf.NAME, conf.TZ)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Error when connecting to database", err.Error())
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Error when pinging database", err.Error())
	}

	goquDB := goqu.New("postgres", db)

	return goquDB, db
}