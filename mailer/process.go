package mailer

import (
	"strconv"
	"net/smtp"
	"jobserver/database"
	"log"
)

type Process struct {
	ch chan Message
	db *database.Mailer
}

func (t *Process) Init( db *database.Object ) (ch chan Message) {
	t.ch = make( chan Message )
	t.db = db.Mailer

	return t.ch
}

// TODO A kiküldendő maileket előbb valami queue-ba kéne rakni, amit mentünk valami storageba, és a leveleket ezután kiküldeni, hogy elhalás esetén sem legyen kiküldetlen levél
func (t *Process) Run() {

	log.Println("Mailer: loading")
	if error := t.db.Load("mailer.yaml"); error != nil {
		log.Println("Mailer: ", error)
		return
	}

	log.Println("Mailer: start")
	for message := range t.ch {

		log.Println("Mailer: send..")
		error := t.send(message.Recipient, message.Body)
		if error != nil {
			log.Println("Mailer: ", error )
		}

		log.Println("Mailer: ..done")
	}

	log.Println("Mailer: stop");
}

func (t *Process) send(recipient string, body string) error {
	return smtp.SendMail(t.db.Config.Host+":"+strconv.Itoa(t.db.Config.Port), smtp.PlainAuth("", t.db.Config.Username, t.db.Config.Password, t.db.Config.Host), t.db.Config.From, []string{recipient}, []byte(body))
}
