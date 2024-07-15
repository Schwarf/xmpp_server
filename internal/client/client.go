package client

import (
	"context"
	"encoding/xml"
	"log"
	"net"
	"xmpp_server/internal/constants"

	"mellium.im/xmlstream"
	"mellium.im/xmpp"
	"mellium.im/xmpp/jid"
	"mellium.im/xmpp/stanza"
)

type Message struct {
	stanza.Message
	Body string `xml:"body"`
}

func StartClient() {
	connection, err := net.Dial("tcp", "localhost:5222")
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer connection.Close()

	clientJID := jid.MustParse("client@example.com")

	session, err := xmpp.NewClientSession(
		context.TODO(),
		clientJID,
		connection,
	)
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	message := Message{
		Message: stanza.Message{
			To:   jid.MustParse("server@example.com"),
			Type: stanza.NormalMessage,
		},
		Body: "Hello, XMPP!",
	}

	err = session.Send(context.TODO(), message.Wrap(nil))
	if err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	log.Println("Message sent, waiting for reply...")

	err = session.Serve(xmpp.HandlerFunc(func(t xmlstream.TokenReadEncoder, start *xml.StartElement) error {
		switch start.Name.Local {
		case constants.MessageString:
			var reply Message
			err := xml.NewTokenDecoder(t).DecodeElement(&reply, start)
			if err != nil {
				return err
			}
			log.Printf("Received message: %v", reply.Body)
		case constants.PresenceString:
			var pres stanza.Presence
			err := xml.NewTokenDecoder(t).DecodeElement(&pres, start)
			if err != nil {
				return err
			}
			log.Println("Received presence from:", pres.From)
		case constants.IQString:
			var iq stanza.IQ
			err := xml.NewTokenDecoder(t).DecodeElement(&iq, start)
			if err != nil {
				return err
			}
			log.Println("Received IQ from:", iq.From)
		default:
			log.Printf("Received unknown stanza: %v", start.Name.Local)
		}
		return nil
	}))
	if err != nil {
		log.Fatalf("Failed to serve session: %v", err)
	}
}
