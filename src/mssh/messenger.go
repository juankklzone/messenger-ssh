package mssh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

const (
	//deliverURL es el formato de la url para enviar mensajes
	deliverURL = "https://graph.facebook.com/v2.6/me/messages?access_token=%s"
	//ayuda muestra los comandos que pueden realizarse con el servidor
	ayuda = `
Lista de comandos:

start ssh <usuario> <direccion> [puerto]
Se comunica a un servidor ssh usuario@direccion. A partir de este momento
los comandos que introduzcas serán procesados por el servidor remoto. 

close
Cierra la conexión ssh

help
Muestra este comando de ayuda 
`
)

var (
	//url ya contiene el token de la página despues de pasar por la función init
	url string

	//PageAuth es la estrucutra que guarda los tokens de la página de Facebook
	PageAuth = Auth{
		// VerifyToken contiene el token para verificar la autenticidad del servicio web, es definido en la interfaz del API de Messenger
		VerifyToken: "verify_Alepht",
		//PageToken contiene el token generado en el API de Messenger para utilizar una págian de Facebook
		PageToken: "EAALVTDYz2rcBAD2NvvkJMo9d987bVfbMaXIC35d2DHtfnwLAFkQbtBfSacLBA5ch94prbcL9ZAXsDe72UAiXMUaahOyZB39dXgYdE8eDNt0gXlK6Ag3YbJInHCUKM78THlVh0k8F2wU5PAwWyEGzSZCq4MkLGq09OWY2bgjwAZDZD",
	}
)

type (
	//Auth contiene los Tokens del API de Messenger
	Auth struct {
		VerifyToken string
		PageToken   string
	}

	//Data es el JSON recibido del API de Messenger al recibir un mensaje
	Data struct {
		Object  string  `json:"object"`
		Entries []Entry `json:"entry"`
	}

	//Entry contiene información sobre los mensajes
	Entry struct {
		ID         string      `json:"id"`
		Time       int         `json:"time"`
		Messagings []Messaging `json:"messaging"`
	}

	//Messaging es utilizado para saber información sobre los mensajes
	Messaging struct {
		Sender    Sender    `json:"sender"`
		Recipient Recipient `json:"recipient"`
		Timestamp int       `json:"timestamp"`
		Message   Message   `json:"message"`
	}

	//Sender es el que envía el mensaje
	Sender struct {
		ID string `json:"id"`
	}

	//Recipient contiene información sobre el receptor del mensaje
	Recipient struct {
		ID string `json:"id"`
	}

	//Message es la información del mensaje
	Message struct {
		Mid  string `json:"mid,omitempty"`
		Seq  int    `json:"seq,omitempty"`
		Text string `json:"text,omitempty"`
	}

	//DeliverMessage es utilizado para envíar mensajes desde el servidor
	DeliverMessage struct {
		Recipient Recipient `json:"recipient"`
		Message   Message   `json:"message"`
	}
)

//init es utilizado para crear el url donde se enviarán los mensajes
func init() {
	url = fmt.Sprintf(deliverURL, PageAuth.PageToken) //Url donde se envían los mensajes
}

//sendMessage es el encargado de responder al usuario
func sendMessage(id string, message string) {
	if len(message) > 0 { //Validación para no enviar mensajes vacíos
		log.Println(id, message)
		dm := DeliverMessage{
			Recipient: Recipient{ID: id},
		}
		for _, text := range blockText(message) { //Se parte el mensaje en bloques de 320 caracteres
			dm.Message.Text = text
			message, err := json.Marshal(&dm) //Convirtiendo la estructura a JSON para enviarlo
			if err != nil {
				log.Println(id, "Error al codificar mensaje de envio:", err)
				return
			}
			resp, err := http.Post(url, "application/json", bytes.NewBuffer(message)) //Petición POST a la url para el envío de mensajes
			if err != nil {
				log.Println(id, "Error al enviar respuesta:", err)
			}
			if resp.StatusCode != http.StatusOK { //Sí no hubo una respuesta 200 entonces ocurrió un error
				log.Println(id, "Status equivocado de respuesta:", resp.Status)
			}
		}
	}
}

//HanddleMessage se encaarga de hacer la función dependiendo el mensaje recibido
func HanddleMessage(m Messaging) {
	checkmsg := strings.ToLower(strings.TrimSpace(m.Message.Text)) //Se convierte a minusculas y se eliminan los espacios
	checkmsg = strings.Split(checkmsg, " ")[0]                     //Se obtiene la primera palabra del mensaje para saber como tratarlo
	switch checkmsg {
	case "start": //Comando para iniciar una conexión SSH
		sendMessage(m.Sender.ID, "Conectando...")
		err := startSession(m) //Se inicia una conexión del usuario
		if err != nil {
			sendMessage(m.Sender.ID, "No se pudo conectar al servidor")
			log.Println(m.Sender.ID, "No se pudo establecer la conexión:", err)
			return
		}
		sendMessage(m.Sender.ID, "Conexión realizada")
	case "close": //Comando para cerrar conexiones activas
		sendMessage(m.Sender.ID, "Cerrando sesión....")
		err := closeSession(m) //Se cierra la conexión del usuario
		message := "Sesión finalizada"
		if err != nil {
			if err == errNoSesion {
				log.Println(m.Sender.ID, "Error al cerrar sesión:", err)
				message = "No existe una sesión activa"
			} else {
				log.Println(m.Sender.ID, "Error al cerrar sesión:", err)
				message = "Error al cerrar sesión"
			}
		}
		sendMessage(m.Sender.ID, message)
	case "help": //Comando para mostrar la ayuda
		enviarAyuda(m.Sender.ID)
	default: //Si no es un comando definido quiere decir que es un comando para el servidor remoto
		if checkmsg == "vi" || checkmsg == "nano" || checkmsg == "vim" { //No se puede ejecutar editores como vi,nano o vim
			sendMessage(m.Sender.ID, "No se pudo ejecutar comando")
			return
		}
		isCdCommand := checkmsg == "cd" //Se da un trato especial a los comandos de cd
		if isCdCommand {                //Si el comando contiene cd
			//Se dirige a la ruta guardada y se agrega un pwd para saber en que carpeta quedó
			m.Message.Text = fmt.Sprintf("cd %s && %s && pwd", getPath(m.Sender.ID), m.Message.Text)
		} else {
			//Se dirige a la ruta guardada y se ejecuta el comando
			m.Message.Text = fmt.Sprintf("cd %s && %s", getPath(m.Sender.ID), m.Message.Text)
		}
		result, err := sendCommand(m) //Se manda el comando a ejecutar al servidor remoto
		if err != nil {
			sendMessage(m.Sender.ID, "No se pudo ejecutar comando")
			if err == errNoSesion {
				enviarAyuda(m.Sender.ID) //Si no existe sesión se manda la ayuda
			}
			log.Println(m.Sender.ID, "Error al enviar comando:", err)
		} else {
			if isCdCommand { //Si el comando contiene cd se da guarda la carpeta actual para usarla posteriormente con cada comando
				fmt.Println(result)
				result = result[:len(result)-1] //Se obtiene la salida del comando pwd para saber la ruta actual
				updatePath(m.Sender.ID, result) //Se actualiza la ruta
			}
			sendMessage(m.Sender.ID, result) //Se da respuesta al usuario
		}
	}
}

//enviarAyuda se encarga notificarle al usuario que comandos pueden ser utilizados
func enviarAyuda(senderID string) {
	sendMessage(senderID, ayuda)
}

//blockText fragmenta el mensaje en paquetes de 320 caracteres por la restricción del API
func blockText(text string) []string {
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
