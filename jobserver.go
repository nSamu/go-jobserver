package main

import (
	"net/http"
	"jobserver/mailer"
	"jobserver/database"
	"jobserver/worker"
	"jobserver/api"
)

func main() {

	db_object := new( database.Object )
	db_object.Init()

	// mailer
	m_object := new( mailer.Process )
	m_channel := m_object.Init( db_object )

	// worker
	w_object := new( worker.Process )
	w_channel := w_object.Init( db_object, m_channel )

	// api
	a_object := new( api.Process )
	a_object.Init( db_object, w_channel )

	http.HandleFunc("/", a_object.Run )
	http.ListenAndServe(":8080", nil)
}
