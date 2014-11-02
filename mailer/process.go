package mailer

import (
	"strconv"
	"net/smtp"
	"jobserver/database"
	"log"
	"fmt"
	"encoding/base64"
	"net/mail"
	"strings"
)

const LISTENER_COUNT = 2

type Process struct {
	ch   chan Message
	db *database.Mailer
}

// A levélküldő inicializálása
func (t *Process) Init(db *database.Object) chan<- Message {

	// adatbázis elmentése
	t.db = db.Mailer

	// konfigurációs fájl beolvasás
	if error := t.db.Load("mailer.yaml"); error != nil {
		panic("Mailer: can't load the configuration: " + error.Error())
	}

	// TODO adatbázisból kiküldetlen levelek kiküldésre küldése (ezt így valahogy optimálni kell majd, mert ha csak forral bedobáljuk a channelbe, akkor lehet kicsit telítődik egy ideig

	// alapértelmezett futtatás
	return t.Run()
}

// Levélküldés figyelők indítása
func (t *Process) Run() chan<- Message {

	// lezárás előkészítése
	t.Stop()
	t.ch = make(chan Message)

	// figyelők létrehozása
	for i := 0; i < LISTENER_COUNT; i++ {

		// TODO levél elmentése adatbázisba, hogy hiba esetén visszaállítható legyenek a kiküldetlen levelek

		go t.listen(t.ch)
	}

	return t.ch
}
// Futás leállítása, összes levélküldő figyelő megállítása
func (t *Process) Stop() {
	defer func() {
		recover()
	}()

	if t.ch != nil {
		close(t.ch)
	}
}

// Aktuális levél query (channel) lekérdezése
func (t *Process) Channel() chan<- Message {
	return t.ch
}

// Tényleges levélküldő, ami figyeli a belső csatornát
func (t *Process) listen(channel <-chan Message) {
	log.Println("Mailer: listen..");

	for {

		if message, open := <-channel; !open {

			log.Println("Mailer: ..stop listen");
			return;
		} else {

			log.Println("Mailer: send (" + message.Recipient + ")..");
			t.send(message.Recipient, message.Subject, message.Body)
			log.Println("Mailer: ..sent (" + message.Recipient + ")");
		}
	}
}
// Tényleges levélküldés a megadott címzettnek és üzenettel
func (t *Process) send(recipient string, subject string, body string) {

	// host string és authentikációs adatok összeállítása
	host := t.db.Config.Host + ":" + strconv.Itoa(t.db.Config.Port)
	auth := smtp.PlainAuth("", t.db.Config.Username, t.db.Config.Password, t.db.Config.Host)

	// fejléc definiálása
	header := make(map[string]string)
	header["From"] = t.db.Config.From
	header["To"] = recipient
	header["Subject"] = (&mail.Address{ subject, ""}).String()
	header["Subject"] = strings.Trim( header["Subject"], "\" <>")
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	// fejléc összeillesztése a tartalommal
	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

	// levél kiküldése az összeszedett adatokkal
	error := smtp.SendMail(host, auth, t.db.Config.From, []string{recipient}, []byte(message))

	if error != nil {
		log.Println("Mailer: (", error, ") error while sending")
		// TODO levél újraküldése X alkalommal?
	} else {
		// TODO levél törlése az adatbázisból, mert ugye már elment
	}
}
