package worker

import (
	"jobserver/database"
	"jobserver/mailer"
);

type Process struct {
	ch chan Message
	mch chan mailer.Message

	db *database.Object
}

type Message struct {
	Id, Callback string
}

func (t *Process) Init( db *database.Object, mch chan mailer.Message ) ( ch chan Message ) {
	t.ch = make( chan Message )
	t.mch = mch
	t.db = db

	return t.ch
}

func (t *Process) Run() {}
