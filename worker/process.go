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

func (t *Process) Init( db *database.Object, mch chan mailer.Message ) ( ch chan Message ) {
	t.ch = make( chan Message )
	t.mch = mch
	t.db = db

	go t.Run()

	return t.ch
}

func (t *Process) Run() {

	for message := range t.ch {
		go t.execute( message )
	}
}

func (t *Process) execute( data Message ) {

}
