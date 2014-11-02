package database

import (
	"io/ioutil"
	"gopkg.in/yaml.v2/yaml"
)

type Worker struct {
	list map[string]Job
}

type Job struct {
	Title, Description, Command string
	Enabled bool
	Dependencies []string
	Email string
}

func (t *Worker) Load( file string ) error {

	data, error := ioutil.ReadFile(file)
	if error == nil {

		t.list = make(map[string]Job)
		error = yaml.Unmarshal(data, &t.list )
	}

	return error
}
func (t *Worker) Get( name string ) ( *Job, bool ) {

	if t.list == nil {
		return nil, false
	}

	job, exist := t.list[ name ]
	return &job, exist
}

func (t *Worker) GetList() map[string]Job {
	return t.list
}
