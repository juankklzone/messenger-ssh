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
	f, err := os.OpenFile("./errors.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic("No se pudo abrir el archivo log: " + err.Error())
	}
	log.SetOutput(f)
}

func main() {
	http.HandleFunc("/", handdler)
	http.ListenAndServe(":8000", nil)
}

func handdler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		verify(res, req)
	case http.MethodPost:
		recieve(res, req)
	default:
		log.Println("Metodo no permitido: ", req.Method)
	}
}

func verify(res http.ResponseWriter, req *http.Request) {
	if mssh.PageAuth.VerifyToken == req.FormValue("hub.verify_token") {
		res.Write([]byte(req.FormValue("hub.challenge")))
		return
	}
	res.Write([]byte("Error,wrong validation token"))
}

func recieve(res http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Println(err)
		return
	}
	var d mssh.Data
	err = json.Unmarshal(body, &d)
	if err != nil {
		log.Println(err)
		return
	}
	for _, m := range d.Entries[0].Messagings {
		if len(m.Message.Text) > 0 {
			mssh.HanddleMessage(m)
		}
	}
}
