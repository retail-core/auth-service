package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/retail-core/auth-service/internal/db/generated"
)

func Connect(dataSource string) (*sql.DB, *db.Queries) {
	_db, err := sql.Open("postgres", dataSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	if err := _db.Ping(); err != nil {
		log.Fatal("cannot ping db:", err)
	}

	log.Println("connected to db successfully")
	q := db.New(_db)
	return _db, q
}