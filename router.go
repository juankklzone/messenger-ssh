package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mssh"
	"net/http"
)

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
	body, _ := ioutil.ReadAll(req.Body)
	var d mssh.Data
	err := json.Unmarshal(body, &d)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, m := range d.Entries[0].Messagings {
		if m.Message.Text != "" {
			mssh.HanddleMessage(m)
		}
	}
}
