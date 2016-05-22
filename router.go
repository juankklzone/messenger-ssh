package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const deliverURL = "https://graph.facebook.com/v2.6/me/messages?access_token=%s"

var a = Auth{
	VerifyToken: "verify_Alepht",
	PageToken:   "EAALVTDYz2rcBAD2NvvkJMo9d987bVfbMaXIC35d2DHtfnwLAFkQbtBfSacLBA5ch94prbcL9ZAXsDe72UAiXMUaahOyZB39dXgYdE8eDNt0gXlK6Ag3YbJInHCUKM78THlVh0k8F2wU5PAwWyEGzSZCq4MkLGq09OWY2bgjwAZDZD",
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
	}
}

func verify(res http.ResponseWriter, req *http.Request) {
	if a.VerifyToken == req.FormValue("hub.verify_token") {
		res.Write([]byte(req.FormValue("hub.challenge")))
		return
	}
	res.Write([]byte("Error,wrong validation token"))
}

func recieve(res http.ResponseWriter, req *http.Request) {
	body, _ := ioutil.ReadAll(req.Body)
	var d Data
	err := json.Unmarshal(body, &d)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, m := range d.Entries[0].Messagings {
		if m.Message.Text != "" {
			handdleMessage(m)
		}
	}
}
