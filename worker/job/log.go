package job

import "time"

const LOG_DIRECTORY = "log/"
type Log struct {
	job string
	index, result int

	output, error string
	time_start, time_stop time.Time
}

func (t *Log) Load( job *Object, index int ) *Log {
	return t
}
func (t *Log) Start( job *Object ) *Log {
	t.time_start = time.Now()

	return t
}
func (t *Log) Stop( result int, output string, error error ) *Log {
	t.result = result
	t.output = output

	if error != nil {
		t.error = error.Error()
	} else {
		t.error = ""
	}

	t.time_stop = time.Now()
	return t
}
