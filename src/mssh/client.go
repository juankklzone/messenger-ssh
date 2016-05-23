package mssh

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/ssh"
)

//User contiene la configuración para la conexión SSH y el id que lo representa en Messenger
type User struct {
	id   string
	conn *ssh.Client
}

var (
	mapaUsuarios   map[string]User
	mapaPermitidos map[string][]string
	auth           ssh.AuthMethod
	ruta           = os.Getenv("SSH_KEY")
)

func init() {
	mapaUsuarios = make(map[string]User)
	mapaPermitidos = make(map[string][]string)
	mapaPermitidos["1026748750723907"] = []string{"alepht.com", "alepht", "104.236.30.229"}  //Juan
	mapaPermitidos["10205869268711621"] = []string{"alepht.com", "alepht", "104.236.30.229"} //Mario
	auth = publicKeyFile(ruta)
}

func startSession(m Messaging) (err error) {
	ips, val := mapaPermitidos[m.Sender.Id]
	if val {
		var user, ip, port string
		fmt.Sscanf(m.Message.Text, "start ssh %s %s %s", &user, &ip, &port)
		if allowIP(ip, ips) {
			u := User{
				id: m.Sender.Id,
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
			mapaUsuarios[u.id] = u
			return
		}
	}
	err = errors.New("No se tiene autorización")
	return
}

func closeSession(m Messaging) (err error) {
	err = mapaUsuarios[m.Sender.Id].conn.Close()
	delete(mapaUsuarios, m.Sender.Id)
	return
}

func sendCommand(m Messaging) (result string, err error) {
	usr := mapaUsuarios[m.Sender.Id]
	if usr.conn == nil {
		err = errors.New("No hay una sesión iniciada")
		return
	}
	fmt.Println("Comando a enviar: ", m.Message.Text)
	session, _ := usr.conn.NewSession()
	data, err := session.CombinedOutput(m.Message.Text)
	result = string(data)
	session.Close()
	return
}

func allowIP(ip string, ips []string) bool {
	for i := range ips {
		if ips[i] == ip {
			return true
		}
	}
	return false
}

func publicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	checkErr(err)
	key, err := ssh.ParsePrivateKey(buffer)
	checkErr(err)
	return ssh.PublicKeys(key)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
