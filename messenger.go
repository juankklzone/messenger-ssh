package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

//Auth contiene los Tokens del API de Messenger
type Auth struct {
	VerifyToken string
	PageToken   string
}

//Data es el JSON recibido del API de Messenger al recibir un mensaje
type Data struct {
	Object  string  `json:"object"`
	Entries []Entry `json:"entry"`
}

//Entry contiene información sobre los mensajes
type Entry struct {
	Id         string      `json:"id"`
	Time       int         `json:"time"`
	Messagings []Messaging `json:"messaging"`
}

//Messaging es utilizado para saber información sobre los mensajes
type Messaging struct {
	Sender    Sender    `json:"sender"`
	Recipient Recipient `json:"recipient"`
	Timestamp int       `json:"timestamp"`
	Message   Message   `json:"message"`
}

//Sender es el que envía el mensaje
type Sender struct {
	Id string `json:"id"`
}

//Recipient contiene información sobre el receptor del mensaje
type Recipient struct {
	Id string `json:"id"`
}

//Message es la información del mensaje
type Message struct {
	Mid  string `json:"mid,omitempty"`
	Seq  int    `json:"seq,omitempty"`
	Text string `json:"text,omitempty"`
}

//DeliverMessage es utilizado para envíar mensajes desde el servidor
type DeliverMessage struct {
	Recipient Recipient `json:"recipient"`
	Message   Message   `json:"message"`
}

func sendMessage(id string, text string) {
	dm := DeliverMessage{
		Message:   Message{Text: text},
		Recipient: Recipient{Id: id},
	}
	url := fmt.Sprintf(deliverURL, a.PageToken)
	message, err := json.Marshal(&dm)
	if err != nil {
		fmt.Println(err)
		return
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(message))
	if err != nil {
		fmt.Println(err)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println(resp.Status)
	}
}

func handdleMessage(m Messaging) {
	if strings.HasPrefix(m.Message.Text, "start") {
		err := startSession(m)
		if err != nil {
			sendMessage(m.Sender, "No se pudo conectar\nHola")
			return
		}
		sendMessage(m.Sender, "Conexión realizada\nHola")
	}
}
