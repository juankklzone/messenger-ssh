package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/crypto/ssh"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

var (
	user = flag.String("user", os.Getenv("SSH_CLIENT"), "usuario ssh -> $SSH_CLIENT")
	pass = flag.String("pass", os.Getenv("SSH_PASS"), "pass ssh -> $SSH_PASS")
)

func main() {
	flag.Parse()
	config := &ssh.ClientConfig{
		User: *user,
		Auth: []ssh.AuthMethod{
			ssh.Password(*pass),
		},
	}
	conn, err := ssh.Dial("tcp", "localhost:22", config)
	checkErr(err)
	defer conn.Close()
	session, err := conn.NewSession()
	checkErr(err)
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	pipe, err := session.StdinPipe()
	defer pipe.Close()
	tee := io.TeeReader(os.Stdin, pipe)
	go func(r io.Reader) {
		b, err := ioutil.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%s", b)
	}(tee)
	checkErr(err)

	err = session.Start("/bin/zsh")
	fmt.Println("esperando datos...")
	session.Wait()

	/*modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}*/
	/*err = session.RequestPty("xterm", 120, 60, modes)
	checkErr(err)*/
	/*data, err := session.Output("uname -a")
	checkErr(err)
	fmt.Printf("%q\n", data)*/
}
