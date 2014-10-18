package api

import (
	"jobserver/database"
	"jobserver/worker"
	"net/http"
	"fmt"
)

type Process struct {
	wch chan worker.Message
	db *database.Object
}

func (t *Process) Init( db *database.Object, wch chan worker.Message ) {
	t.db = db
	t.wch = wch
}

func (t *Process) Run(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}
