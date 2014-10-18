package database

import (
	"io/ioutil"
	"gopkg.in/yaml.v2/yaml"
)

type Mailer struct {
	Config struct {
		Host     string
		Port     int
		Username, Password string
		From      string
	}
}

func (t *Mailer) Load( file string ) error {

	data, error := ioutil.ReadFile(file)
	if error == nil {
		error = yaml.Unmarshal(data, &t.Config )
	}

	return error
}
func (t *Mailer) Backup() {}
func (t *Mailer) Restore() {}
