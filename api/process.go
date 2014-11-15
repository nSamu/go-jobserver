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

// Kérés kezelés minta
type handler map[string]func( []string, map[string][]string, map[string][]string ) response

// Egy kérés válasza
type response struct {
	data interface {}
	error error
	status int
}

// API által megszabott job szerkezet
type jobItem struct {
	Name, Title, Description string
}
// API által megszabott futási napló szerkezet
type logItem struct {
	Id, Result int
	Stdin, Stdout, Stderr string
	Time int64
}
// API által megszabott futási napló rövid szerkezet
type logItemShort struct {
	Id, Result int
	Time int64
}

type Process struct {
	wch chan<- worker.Message
	db *database.Object

	handlers handler
}

// Egyszerű API kezelő inicailizálása
func (t *Process) Init( db *database.Object, wch chan<- worker.Message ) {
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

// Visszaad egy listát az elérhető jobokról. Minden jobról elég csak minimális
// információt adni (name, olvasható név, leírás).
func (t *Process) requestList( data []string, get map[string][]string, post map[string][]string ) response {

	// job lista lekérdezése
	list, i := t.db.Worker.GetList(), 0

	// lekért job lista adatainak átmásolása a kívánt formátumban
	result := make( []jobItem, 0, len(list) )
	for name, job := range list {
		result[i] = jobItem{ name, job.Title, job.Description }
		i++
	}

	return response{ result, nil, http.StatusOK }
}
// Visszaad minden információt a jobról, és egy rövid összegzést (futás száma,
// ideje, eredménye) a legutóbbi 20 futásról időrend szerint fordítva rendezve
// (legújabb elől).
func (t *Process) requestData( data []string, get map[string][]string, post map[string][]string ) response {

	// lekérdezendő job adatainak kigyűjtése
	if job := t.db.Worker.Get( data[0] ); job == nil {
		return response{ nil, errors.New("Job Not Found"), http.StatusNotFound }
	} else {

		tmp := t.db.GetRunList( data[0], 0, 20 )
		logs := make( []logItemShort, 0, len( tmp ) )
		log.Println( logs )
		for _, l := range tmp {
			logs = append( logs, logItemShort{
				Id: l.Id,
				Result: l.Result,
				Time: l.Stop.Unix() - l.Start.Unix(),
			} )
		}

		return response{ struct{
			Title, Description, Command string
			Dependencies []string
			Email string
			Log []logItemShort
		}{
			job.Title,
			job.Description,
			job.Command,
			job.Dependencies,
			job.Email,
			logs,
		}, nil, http.StatusOK }
	}

	return response{ data, nil, http.StatusOK }
}
// Egy adott job lefutásával kapcsolatos adatokat ad vissza (stdin ha volt,
// eredmény, shell parancs stdout + stderr).
func (t *Process) requestLog( data []string, get map[string][]string, post map[string][]string ) response {

	// lekérdezendő job adatainak kigyűjtése
	if tmp := t.db.Worker.Get( data[0] ); tmp == nil {
		return response{ nil, errors.New("Job Not Found"), http.StatusNotFound }
	} else {

		index, _ := strconv.Atoi( data[1] )
		if tmp := t.db.GetRun( data[0], index ); tmp == nil {
			return response{ nil, errors.New("Run Not Found"), http.StatusNotFound }
		} else {

			return response{ logItem{
				Id: tmp.Id,
				Result: tmp.Result,
				Stdin: tmp.Stdin,
				Stdout: tmp.Stdout,
				Stderr: tmp.Stderr,
				Time: tmp.Stop.Unix()-tmp.Start.Unix(),
			}, nil, http.StatusOK }
		}
	}
}
// Előjegyez egy jobot lefutásra. Opcionális get paraméter a callback, amire
// egy POST hívást csinál (a futás által generált stdout és stderr kombinálva
// a POST adat), ha vége a jobnak. A POST-ként küldött adatot egy az egyben a
// job által definiált shell parancs stdin-jére másolja át.
func (t *Process) requestRun( data []string, get map[string][]string, post map[string][]string ) response {

	// futtatandó job gyorsellenőrzése
	if tmp := t.db.Worker.Get( data[0] ); tmp == nil {
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
