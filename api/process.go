package api

import (
	"jobserver/database"
	"jobserver/worker"
	"net/http"
	"strconv"
	"strings"
	"log"
	"encoding/json"
	"errors"
)

type handler map[string]func( []string, map[string][]string, map[string][]string ) response

type Process struct {
	wch chan worker.Message
	db *database.Object

	handlers handler
}

// Egyszerű API kezelő inicailizálása
func (t *Process) Init( db *database.Object, wch chan worker.Message ) {
	t.db = db
	t.wch = wch

	// kérés feldolgozók definiálása: <METHOD>-<PATHLEN>
	t.handlers = make( handler )
	t.handlers["GET-0"] = t.requestList
	t.handlers["POST-1"] = t.requestRun
	t.handlers["GET-1"] = t.requestData
	t.handlers["GET-2"] = t.requestLog
}

// Szerver indítás
func (t *Process) Run() {
	http.HandleFunc("/", t.serve )
	http.ListenAndServe(":8080", nil)
}

// Egy kérés kiszolgálása
func (t *Process) serve(writer http.ResponseWriter, request *http.Request) {

	// hívott url darabok feldolgozása
	paths := []string{}
	if path := strings.Trim( request.URL.Path, "/" ); path != "" {
		paths = strings.Split( path, "/" )
	}

	// darabok és metódus alapján kezelő választás és futtatás
	var r response
	handler_name := request.Method + "-" + strconv.Itoa( len( paths ) )
	if handler, exist := t.handlers[ handler_name ]; exist {
		r = handler( paths, request.URL.Query(), request.PostForm )
	} else {
		r = response{ nil, errors.New( http.StatusText( http.StatusNotFound ) ), http.StatusNotFound }
	}

	// eredmény feldolgozása
	log.Println( "API: request..", r )
	if r.error != nil {
		http.Error( writer, r.error.Error(), r.status )
	} else {

		if body, error := json.Marshal( r.data ); error != nil {
			http.Error( writer, http.StatusText( http.StatusInternalServerError ), http.StatusInternalServerError )
		} else {
			writer.WriteHeader( r.status )
			writer.Write( body )
		}
	}
}

// Visszaad egy listát az elérhető jobokról. Minden jobról elég csak minimális információt adni (name, olvasható név, leírás).
func (t *Process) requestList( data []string, get map[string][]string, post map[string][]string ) response {
	return response{ data, nil, http.StatusOK }
}
// Visszaad minden információt a jobról, és egy rövid összegzést (futás száma, ideje, eredménye) a legutóbbi 20 futásról időrend szerint fordítva rendezve (legújabb elől).
func (t *Process) requestData( data []string, get map[string][]string, post map[string][]string ) response {
	return response{ data, nil, http.StatusOK }
}
// Egy adott job lefutásával kapcsolatos adatokat ad vissza (stdin ha volt, eredmény, shell parancs stdout + stderr).
func (t *Process) requestLog( data []string, get map[string][]string, post map[string][]string ) response {
	return response{ data, nil, http.StatusOK }
}
// Előjegyez egy jobot lefutásra. Opcionális get paraméter a callback, amire egy POST hívást csinál (a futás által generált stdout és stderr kombinálva a POST adat), ha vége a jobnak.
// A POST-ként küldött adatot egy az egyben a job által definiált shell parancs stdin-jére másolja át.
func (t *Process) requestRun( data []string, get map[string][]string, post map[string][]string ) response {

	// futtatandó job gyorsellenőrzése
	if !t.db.Worker.Exist( data[0] ) {
		return response{ nil, errors.New("Job Not Found"), http.StatusNotFound }
	}

	// callback url begyűjtése
	var callback string
	if tmp, exist := get["callback"]; exist {
		callback = strings.Join( tmp, "" )
	}

	// adatok elküldése a Worker-nek
	t.wch <- worker.Message{
		Id: data[0],
		Callback: callback,
		Data: post,
	}

	return response{ struct{}{}, nil, http.StatusAccepted }
}
