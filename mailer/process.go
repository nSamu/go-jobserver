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

	go t.Run()

	// TODO adatbázisból kiküldetlen levelek kiküldésre küldése (ezt így vaahogy optimálni kell majd, mert ha csak forral bedobáljuk a channelbe, akkor lehet kicsit telítődik egy ideig

	return t.ch
}

func (t *Process) Run() {

	log.Println("Mailer: loading")
	if error := t.db.Load("mailer.yaml"); error != nil {
		log.Println("Mailer: ", error)
		return
	}

	log.Println("Mailer: start")
	for message := range t.ch {
		// TODO levél elmentése adatbázisba, hogy hiba esetén visszaállítható legyenek a kiküldetlen levelek

		go t.send(message.Recipient, message.Body)
	}

	log.Println("Mailer: stop")
}

func (t *Process) send(recipient string, body string) {
	log.Println("Mailer: send..")

	error := smtp.SendMail(t.db.Config.Host+":"+strconv.Itoa(t.db.Config.Port), smtp.PlainAuth("", t.db.Config.Username, t.db.Config.Password, t.db.Config.Host), t.db.Config.From, []string{recipient}, []byte(body))

	// TODO levél törlése az adatbázisból, mert ugye már elment

	if error != nil {
		log.Println("Mailer: ", error )

		// TODO levél újraküldése X alkalommal?
	}

	log.Println("Mailer: ..done")
}
