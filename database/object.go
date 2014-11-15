package database

import (
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
)

type Object struct {
	*Worker
	*Mailer
}

func (t *Object) Init() *Object {
	t.Worker = new( Worker )
	t.Mailer = new( Mailer )

	// adatbázis létrehozása/megnyitása a későbbi használatra (a Worker illetve a Mailer fogja használni)
	db, error := sql.Open("sqlite3", "./jobserver.db")
	if error != nil {
		panic( error )
	}

	// FIXME ez az átadási mód még csiszolandó
	t.Worker.database, t.Mailer.database = db, db
	return t
}
