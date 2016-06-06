package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"mssh"
	"net/http"
	"os"
)

func init() {
	//Archivo log para guardar errores
	f, err := os.OpenFile("./errors.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic("No se pudo abrir el archivo log: " + err.Error())
	}
	log.SetOutput(f)
}

func main() {
	http.HandleFunc("/", handdler)    //Handler para recibir peticiones
	http.ListenAndServe(":8000", nil) //Corriendo en el puerto 8000
}

//handdler se divide en 2 métodos: GET y POST
func handdler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet: //Cuando existe una petición GET es utilizada para verificar las credenciales del API
		verify(res, req)
	case http.MethodPost: //Cuando el API manda un mensaje el método POST es el encargado de recibirlo y tratarlo
		recieve(res, req)
	default:
		log.Println("Metodo no permitido:", req.Method)
	}
}

func verify(res http.ResponseWriter, req *http.Request) {
	if mssh.PageAuth.VerifyToken == req.FormValue("hub.verify_token") { //Verificando que el token sea el mismo
		//Si el token es el mismo se regresa el token creado
		res.Write([]byte(req.FormValue("hub.challenge")))
		return
	}
	res.Write([]byte("Error,wrong validation token"))
}

func recieve(res http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body) //Leemos el cuerpo de la petición
	if err != nil {
		log.Println(err)
		return
	}
	var d mssh.Data
	err = json.Unmarshal(body, &d) //Convertimos el JSON recibido a una estructura Data
	if err != nil {
		log.Println(err)
		return
	}
	for _, m := range d.Entries[0].Messagings { //Por cada mensaje se da un trato
		if len(m.Message.Text) > 0 { //Si el contenido del mensaje es mayor a 0 entonces se maneja el mismo
			mssh.HanddleMessage(m)
		}
	}
}
