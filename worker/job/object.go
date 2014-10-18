package job

import (
	"os/exec"
	"syscall"
	"strings"
	"sync"
)

type Object struct {
	Command string
	Dependencies []string
	Email string
}

func (t *Object) Run( list *List, deferred *sync.WaitGroup ) ( l *Log ) {
	defer func() {
		if deferred != nil {
			deferred.Done()
		}
	}()

	t.execute( t.Command )
	return
}

func (t *Object) prepare( list *List ) ( wg *sync.WaitGroup, error error ) {

	wg = new( sync.WaitGroup )
	wg.Add( len( t.Dependencies ) )
	for _, dependency := range t.Dependencies {
		if j, ok := (*list)[ dependency ]; ok {
			go func() {
				j.Run( list, wg )
			}()
		}
	}

	return
}

func (t *Object) execute( command string ) (result int, output string, error error) {

	var buffer []byte
	parts := strings.Fields( command )
	buffer, error = exec.Command( parts[0], parts[1:]... ).Output()
	output = string( buffer )

	result = 0
	if error != nil {

		result = -1
		if eerror, ok := error.(*exec.ExitError); ok {
			if status, ok := eerror.Sys().(syscall.WaitStatus); ok {
				result = status.ExitStatus();
			}
		}
	}

	return
}
