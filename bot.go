package main

import (
	"fmt"
	"net/http"

	"github.com/frodsan/fbot"
)

func main() {
	bot := fbot.NewBot(fbot.Config{
		AccessToken: "EAALVTDYz2rcBACQkrtrSon8eNXZCtJl40HA1bSwfKpOMV977JFsB7EXuD3hrYxnugW9s837KSxe5w6JscQSwMsbjPZA62ZBIqyyAl2ArsmDlJtxnfEGREvffwJBsefbxs24jyPJTyhdYLvmC65CU7YT8jtyoJMeEL4lqmlGXQZDZD",
		AppSecret:   "2032fab87b7b8197b430cf5a65c9b639",
		VerifyToken: "0b8baf8b94e97c2416584afc2a8e9016",
	})

	bot.On(fbot.EventMessage, func(event *fbot.Event) {
		fmt.Println(event.Message.Text)

		bot.Deliver(fbot.DeliverParams{
			Recipient: event.Sender,
			Message: &fbot.Message{
				Text: event.Message.Text,
			},
		})
	})

	http.Handle("/bot", fbot.Handler(bot))

	http.ListenAndServe(":8000", nil)
}
