package mssh

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/ssh"
)

//User contiene la configuración para la conexión SSH y el id que lo representa en Messenger
type User struct {
	id       string
	conn     *ssh.Client
	session  *ssh.Session
	writeBuf *bytes.Buffer
	readBuf  *bytes.Buffer
}

var (
	mapaUsuarios map[string]User
	auth         ssh.AuthMethod
	ruta         = os.Getenv("SSH_KEY")
	//user = flag.String("user", os.Getenv("SSH_CLIENT"), "usuario ssh -> $SSH_CLIENT")
	//ruta = flag.String("archivo", os.Getenv("SSH_PUBLIC_KEY"), "archivo con llave pública $SSH_PUBLIC_KEY")
	//pass = flag.String("pass", os.Getenv("SSH_PASS"), "pass ssh -> $SSH_PASS")
)

func init() {
	mapaUsuarios = make(map[string]User)
	auth = publicKeyFile(ruta)
}

func startSession(m Messaging) (err error) {
	var user, ip, port string
	fmt.Sscanf(m.Message.Text, "start ssh %s %s %s", &user, &ip, &port)
	u := User{
		id:       m.Sender.Id,
		writeBuf: new(bytes.Buffer),
		readBuf:  new(bytes.Buffer),
	}
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			auth,
		},
	}
	if port == "" {
		port = "22"
	}
	url := ip + ":" + port
	fmt.Println(url)
	u.conn, err = ssh.Dial("tcp", url, config)
	if err != nil {
		return
	}
	u.session, err = u.conn.NewSession()
	if err != nil {
		return
	}
	u.session.Stdout = u.readBuf
	u.session.Stdin = u.writeBuf
	mapaUsuarios[u.id] = u
	go func() { u.session.Wait() }()
	return
}

func closeSession(m Messaging) (err error) {
	err = mapaUsuarios[m.Sender.Id].session.Close()
	delete(mapaUsuarios, m.Sender.Id)
	return
}

func sendCommand(m Messaging) (result string, err error) {
	usr := mapaUsuarios[m.Sender.Id]
	if usr.session == nil {
		err = errors.New("no hay una sesión iniciada")
		return
	}
	_, err = usr.writeBuf.Write([]byte(m.Message.Text))
	usr.writeBuf.Reset()
	result = usr.readBuf.String()
	usr.readBuf.Reset()
	return
}

// func main() {
// 	flag.Parse()
// 	config := &ssh.ClientConfig{
// 		User: *user,
// 		Auth: []ssh.AuthMethod{
// 			publicKeyFile(*ruta),
// 		},
// 	}
// 	conn, err := ssh.Dial("tcp", "alepht.com:22", config)
// 	checkErr(err)
// 	defer conn.Close()
// 	session, err := conn.NewSession()
// 	checkErr(err)
// 	defer session.Close()

// 	session.Stdout = os.Stdout
// 	session.Stderr = os.Stderr
// 	pipe, err := session.StdinPipe()
// 	defer pipe.Close()
// 	tee := io.TeeReader(os.Stdin, pipe)
// 	//Pipe entre Stdin local y Stdin de la sesión ssh
// 	leerDatos := func(r io.Reader) {
// 		b, err := ioutil.ReadAll(r)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		fmt.Printf("%s", b)
// 	}
// 	go leerDatos(tee)
// 	checkErr(err)

// 	modes := ssh.TerminalModes{
// 		ssh.ECHO:          0,     // disable echoing
// 		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
// 		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
// 	}
// 	err = session.RequestPty("xterm", 120, 180, modes)
// 	checkErr(err)
// 	if err := session.Shell(); err != nil {
// 		panic(err)
// 	}
// 	err = session.Wait()
// 	fmt.Println("finalizando sesión con error ", err)
// }

func publicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	checkErr(err)

	key, err := ssh.ParsePrivateKey(buffer)
	checkErr(err)
	//fmt.Println(key.PublicKey().Type())
	return ssh.PublicKeys(key)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
