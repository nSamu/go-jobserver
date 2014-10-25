package database

import "jobserver/worker/job"

type Worker struct {

}

func (t *Worker) Get( name string ) *job.Object {
	return new( job.Object )
}
func (t *Worker) GetList() *job.List {
	return new( job.List )
}
func (t *Worker) Exist( name string ) bool {
	return false
}
func (t *Worker) Backup() {}
func (t *Worker) Restore() {}
