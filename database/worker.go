package database

import (
	"io/ioutil"
	"gopkg.in/yaml.v2/yaml"
	"database/sql"
	"time"
)

// Egy job adatai
type Job struct {

	// cím, leírás és futtatandó parancs
	Title, Description, Command string

	// a job futtatható, nem tartalmaz függőségi kört
	Enabled bool

	// függőségek listája, minden elem egy job index
	Dependencies []string

	// email, amire hiba esetén megy a levél
	Email string
}

// Egy job futtatás eredményének adatai
type JobLog struct {

	// futás idje és eredménye
	Id, Result int

	// job indexe
	Job string

	// futtatás bemenet, kimenet és hibakimenet
	Stdin, Stdout, Stderr string

	// befejezéskor hívott url (ha van)
	Callback string

	// kezdő és befejező dátum
	Start, Stop time.Time
}

type Worker struct {
	list map[string]Job

	database *sql.DB
}

// Worker adatbátis inicializálása
func (t *Worker) Init() {

	// konfigurációs fájl beolvasás
	if error := t.Load("worker.yaml"); error != nil {
		panic("Worker: can't load the configuration - " + error.Error())
	}

	// adatbázis létrehozása ha még nem lenne
	command := `CREATE TABLE IF NOT EXISTS worker_log(
		id INTEGER NOT NULL PRIMARY KEY,
		job STRING NOT NULL,
		result INTEGER NOT NULL,
		stdin TEXT,
		stdout TEXT,
		stderr TEXT,
		callback TEXT,
		start INTEGER NOT NULL,
		stop INTEGER NOT NULL,
		time_create INTEGER NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	if _, error := t.database.Exec( command ); error != nil {
		panic("Worker: can't create worker log table - " + error.Error() )
	}
}
// Beállítás fájl betöltése
func (t *Worker) Load( file string ) error {

	data, error := ioutil.ReadFile(file)
	if error == nil {

		t.list = make(map[string]Job)
		error = yaml.Unmarshal(data, &t.list )
	}

	return error
}

// Új futtatási napló hozzáadása
func (t *Worker) AddLog( job string, result int, stdin, stdout, stderr, callback string, start, stop time.Time ) error {

	if tx, error := t.database.Begin(); error == nil {

		command := `INSERT INTO worker_log( job, result, stdin, stdout, stderr, callback, start, stop )
								VALUES( ?, ?, ?, ?, ?, ?, ?, ? )`

		if stmt, error := tx.Prepare( command ); error == nil {

			defer stmt.Close()
			_, error = stmt.Exec( job, result, stdin, stdout, stderr, callback, start.Format( time.RFC3339 ), stop.Format( time.RFC3339 ) )
			if error == nil {
				return tx.Commit()
			}
		}

		return error
	} else {
		return error
	}
}
// Egy futtatási napló lekérdezése
func (t *Worker) GetLog( job string, index int ) *JobLog {

	rows, err := t.database.Query("SELECT result, stdin, stdout, stderr, callback, start, stop FROM worker_log WHERE job = ? AND id = ?", job, index );
	if err == nil {

		defer rows.Close()
		tmp := &JobLog{
			Id: index,
			Job: job,
		}
		var start, stop string

		if rows.Next() && rows.Scan(&tmp.Result, &tmp.Stdin, &tmp.Stdout, &tmp.Stderr, &tmp.Callback, &start, &stop) == nil {

			tmp.Start, _ = time.Parse(time.RFC3339, start)
			tmp.Stop, _ = time.Parse(time.RFC3339, stop)

			return tmp
		}
	}

	return nil
}
// Legutolsó x darab futtatási napló lekérdezése egy adott job-hoz
func (t *Worker) GetLogList( job string, limit int ) []JobLog {

	result := make( []JobLog, 0, limit )
	command :=	`SELECT id, job, result,
											stdin, stdout, stderr,
											callback,
											start, stop
							 FROM worker_log
							 WHERE job = ?
							 ORDER BY stop DESC
							 LIMIT ?`

	if rows, error := t.database.Query( command, job, limit ); error == nil {

		defer rows.Close()
		for rows.Next() {

			tmp := JobLog{}
			var start, stop string
			if rows.Scan( &tmp.Id, &tmp.Job, &tmp.Result, &tmp.Stdin, &tmp.Stdout, &tmp.Stderr, &tmp.Callback, &start, &stop ) == nil {

				// lekérdezett idők feldolgozása, mert stringként vannak tárolva (elvileg int-ként is fel kellene dolgozni mert time-ot nem kezel)
				tmp.Start, _ = time.Parse(time.RFC3339, start)
				tmp.Stop, _ = time.Parse(time.RFC3339, stop)

				result = append(result, tmp)
			}
		}
	}

	return result
}

// Egy job adatainak lekérdezése, index szerint. Nil visszatérés nem létezést jelent
func (t *Worker) Get( name string ) *Job {

	if t.list != nil {
		if job, exist := t.list[ name ]; exist {
			return &job
		}
	}

	return nil
}
// Job lista lekérdezése
func (t *Worker) GetList() map[string]Job {
	return t.list
}
