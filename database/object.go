package database

type Object struct {
	*Worker
	*Mailer
}

func (t *Object) Init() *Object {
	t.Worker = new( Worker )
	t.Mailer = new( Mailer )

	return t
}
