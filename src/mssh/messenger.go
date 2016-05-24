package mssh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

const deliverURL = "https://graph.facebook.com/v2.6/me/messages?access_token=%s"

const ayuda = `
Lista de comandos:

start ssh <usuario> <direccion> [puerto]
Se comunica a un servidor ssh usuario@direccion. A partir de este momento
los comandos que introduzcas serán procesados por el servidor remoto. 

close
Cierra la conexión ssh

help
Muestra este comando de ayuda 
`

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

func sendMessage(id string, message string) {
	if len(message) > 0 {
		log.Println(id, message)
		url := fmt.Sprintf(deliverURL, PageAuth.PageToken)
		dm := DeliverMessage{
			Recipient: Recipient{Id: id},
		}
		for _, text := range blockText(message) {
			dm.Message.Text = text
			message, err := json.Marshal(&dm)
			if err != nil {
				log.Println(id, "Error al codificar mensaje de envio:", err)
				return
			}
			resp, err := http.Post(url, "application/json", bytes.NewBuffer(message))
			if err != nil {
				log.Println(id, "Error al enviar respuesta:", err)
			}
			if resp.StatusCode != http.StatusOK {
				log.Println(id, "Status equivocado de respuesta:", resp.Status)
			}
		}
	}
}

//HanddleMessage se encaarga de hacer la función dependiendo el mensaje recibido
func HanddleMessage(m Messaging) {
	checkmsg := strings.ToLower(strings.TrimSpace(m.Message.Text))
	if strings.HasPrefix(checkmsg, "start") {
		sendMessage(m.Sender.Id, "Conectando...")
		err := startSession(m)
		if err != nil {
			sendMessage(m.Sender.Id, "No se pudo conectar al servidor")
			log.Println(m.Sender.Id, "No se pudo establecer la conexión:", err)
			return
		}
		sendMessage(m.Sender.Id, "Conexión realizada")
	} else if checkmsg == "close" {
		sendMessage(m.Sender.Id, "Cerrando sesión....")
		err := closeSession(m)
		if err != nil {
			log.Println(m.Sender.Id, "Error al cerrar sesión:", err)
		}
		sendMessage(m.Sender.Id, "Sesión finalizada")
	} else if checkmsg == "help" {
		enviarAyuda(m.Sender.Id)
	} else {
		isCdCommand := false
		if strings.HasPrefix(checkmsg, "cd ") {
			isCdCommand = true
		}
		if isCdCommand {
			m.Message.Text += " & pwd"
		} else {
			m.Message.Text = fmt.Sprintf("cd %s && %s", getPath(m.Message.Text), m.Message.Text)
		}
		result, err := sendCommand(m)
		if err != nil {
			sendMessage(m.Sender.Id, "No se pudo ejecutar comando")
			if err == errNoSesion {
				enviarAyuda(m.Sender.Id)
			}
			log.Println(m.Sender.Id, "Error al enviar comando:", err)
		} else {
			if isCdCommand {
				fmt.Println(result)
				//idxStartPath := strings.Index(result, "\n")
				//result = result[idxStartPath+1:]
				updatePath(m.Sender.Id, result)
			} else {
				sendMessage(m.Sender.Id, result)
			}
		}
	}
}

func enviarAyuda(senderId string) {
	sendMessage(senderId, ayuda)
}

func blockText(text string) []string {
	//Límite de caractares a envíar 320
	blocks := make([]string, len(text)/320+1)
	for i := 0; i <= len(text)/320; i++ {
		in := i * 320
		fin := in + 320
		if fin > len(text) {
			fin = len(text)
		}
		if in == fin {
			break
		}
		blocks[i] = text[in:fin]
	}
	return blocks
}
