package mssh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const deliverURL = "https://graph.facebook.com/v2.6/me/messages?access_token=%s"

var PageAuth = Auth{
	VerifyToken: "verify_Alepht",
	PageToken:   "EAALVTDYz2rcBAD2NvvkJMo9d987bVfbMaXIC35d2DHtfnwLAFkQbtBfSacLBA5ch94prbcL9ZAXsDe72UAiXMUaahOyZB39dXgYdE8eDNt0gXlK6Ag3YbJInHCUKM78THlVh0k8F2wU5PAwWyEGzSZCq4MkLGq09OWY2bgjwAZDZD",
}

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
	url := fmt.Sprintf(deliverURL, PageAuth.PageToken)
	message, err := json.Marshal(&dm)
	if err != nil {
		fmt.Println("error al codificar mensaje de envio", err)
		return
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(message))
	if err != nil {
		fmt.Println("error al enviar respesta", err)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println("status equivocado de respuesta", resp.Status)
	}
}

//HanddleMessage se encaarga de hacer la función dependiendo el mensaje recibido
func HanddleMessage(m Messaging) {
	fmt.Println(m.Message.Text)
	if strings.HasPrefix(m.Message.Text, "start") {
		sendMessage(m.Sender.Id, "Conectando...")
		err := startSession(m)
		if err != nil {
			sendMessage(m.Sender.Id, "No se pudo conectar al servidor")
			return
		}
		sendMessage(m.Sender.Id, "Conexión realizada")
	} else if m.Message.Text == "close" {
		sendMessage(m.Sender.Id, "Cerrando sesión....")
		err := closeSession(m)
		if err != nil {
			fmt.Println("error al cerrar sesión", err)
		}
		sendMessage(m.Sender.Id, "sesión finalizada")
	} else {
		result, err := sendCommand(m)
		if err != nil {
			sendMessage(m.Sender.Id, " no se pudo ejecutar comando")
			fmt.Println("error al enviar comando", err)
		} else {
			fmt.Println("resultado a enviar", result)
			sendMessage(m.Sender.Id, result)
		}
	}
}
