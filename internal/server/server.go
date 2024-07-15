package server

import (
	"context"
	"encoding/xml"
	"log"
	"mellium.im/xmlstream"
	"mellium.im/xmpp"
	"mellium.im/xmpp/jid"
	"mellium.im/xmpp/stanza"
	"net"
	"xmpp_server/internal/constants"
)

type Message struct {
	stanza.Message
	Body string `xml:"body"`
}

func StartServer() {
	listener, err := net.Listen("tcp", ":5222")
	if err != nil {
		log.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()
	log.Println("XMPP server started on :5222")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(connection net.Conn) {
	defer connection.Close()

	serverJID := jid.MustParse("server@localhost")
	clientJID := jid.MustParse("client@localhost")

	session, err := xmpp.NewServerSession(
		context.TODO(),
		serverJID,
		clientJID,
		connection,
	)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		return
	}

	err = session.Serve(xmpp.HandlerFunc(func(t xmlstream.TokenReadEncoder, start *xml.StartElement) error {
		switch start.Name.Local {
		case constants.MessageString:
			var message Message
			err := xml.NewTokenDecoder(t).DecodeElement(&message, start)
			if err != nil {
				return err
			}
			log.Printf("Received message: %v", message.Body)
			reply := Message{
				Message: stanza.Message{
					To:   message.From,
					Type: message.Type,
				},
				Body: "Echo: " + message.Body,
			}
			return session.Send(context.TODO(), reply.Wrap(nil))
		case constants.PresenceString:
			var pres stanza.Presence
			err := xml.NewTokenDecoder(t).DecodeElement(&pres, start)
			if err != nil {
				return err
			}
			log.Printf("Received presence from: %v", pres.From)
		case constants.IQString:
			var iq stanza.IQ
			err := xml.NewTokenDecoder(t).DecodeElement(&iq, start)
			if err != nil {
				return err
			}
			log.Printf("Received IQ from: %v", iq.From)
		default:
			log.Printf("Unknown stanza: %v", start.Name.Local)
		}
		return nil
	}))
	if err != nil {
		log.Printf("Failed to serve session: %v", err)
	}
}
