package job

import (
	"io/ioutil"
	"gopkg.in/yaml.v2/yaml"
)

type List map[string]Object

func (t List) Load( file string ) List {
	if data, error := ioutil.ReadFile( file ); error == nil {
		error = yaml.Unmarshal(data, &t );
	}

	return t
}
