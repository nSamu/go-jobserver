package database

import (
	"io/ioutil"
	"gopkg.in/yaml.v2/yaml"
	"database/sql"
)

type Mailer struct {
	Config struct {
		Host     string
		Port     int
		Username, Password string
		From      string
	}

	database *sql.DB
}

func (t* Mailer) Init() {

	// konfigurációs fájl beolvasás
	if error := t.Load("mailer.yaml"); error != nil {
		panic("Mailer: can't load the configuration: " + error.Error())
	}
}

func (t *Mailer) Load( file string ) error {

	data, error := ioutil.ReadFile(file)
	if error == nil {
		error = yaml.Unmarshal(data, &t.Config )
	}

	return error
}
