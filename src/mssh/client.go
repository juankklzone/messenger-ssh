package mssh

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/crypto/ssh"
)

//User contiene la configuración para la conexión SSH y el id que lo representa en Messenger
type User struct {
	id       string
	conn     *ssh.Client
	lastPath string
}

var (
	//mapaUsuarios contiene los usuarios que tienen conexiones
	mapaUsuarios map[string]User
	//mapaPermitidos guarda los hosts permitidos por usuario
	mapaPermitidos map[string][]string
	//auth contiene la autenticación para hacer las conexiones remotas
	auth ssh.AuthMethod
	//ruta es la carpeta donde se encuentran la llave privada SSH
	ruta = os.Getenv("SSH_KEY")
	//errNoSesion es un error definido para indicar que no hay sesión
	errNoSesion = errors.New("No hay una sesión iniciada")
)

func init() {
	mapaUsuarios = make(map[string]User)
	mapaPermitidos = make(map[string][]string)
	//Usuarios con sus hosts permitidos
	mapaPermitidos["1026748750723907"] = []string{"alepht.com", "alepht", "104.236.30.229", "107.170.101.174"}  //Juan
	mapaPermitidos["10205869268711621"] = []string{"alepht.com", "alepht", "104.236.30.229", "107.170.101.174"} //Mario
	//Se obtiene la autenticación utilizando las llaves SSH
	auth = publicKeyFile(ruta)
}

//startSession comienza una conexión SSH
func startSession(m Messaging) (err error) {
	ips, val := mapaPermitidos[m.Sender.ID] //Se verifica que el usuario este en el mapa de los permitidos
	if val {
		var user, ip, port string
		fmt.Sscanf(m.Message.Text, "start ssh %s %s %s", &user, &ip, &port) //Se obtienen los parámetros del comando
		//Se verifica que el host este dentro del mapa del usuario
		if allowIP(ip, ips) {
			//Se crea un nuevo usuario para guardar en el mapa
			u := User{
				id: m.Sender.ID,
			}
			//Configuración necesaria para iniciar la conexión, contiene el usuario y la autenticación
			config := &ssh.ClientConfig{
				User: user,
				Auth: []ssh.AuthMethod{
					auth,
				},
			}
			if port == "" { //Si no hay puerto definido se utiliza el 22 por defecto
				port = "22"
			}
			url := ip + ":" + port                     //Se crea la url usando el host y el puerto
			u.conn, err = ssh.Dial("tcp", url, config) //Se inicia una conexión al servidor remoto
			if err != nil {
				return
			}
			mapaUsuarios[u.id] = u //Guardando el usuario en el mapa
			return
		}
	}
	err = errors.New("No se tiene autorización")
	return
}

//closeSession se encarga de cerrar la conexión
func closeSession(m Messaging) (err error) {
	u := mapaUsuarios[m.Sender.ID] //Se obtiene el usuario del mapa de conexiones
	if u.conn != nil {             //Si existe conexión se cierra
		err = mapaUsuarios[m.Sender.ID].conn.Close()
		delete(mapaUsuarios, m.Sender.ID) //Se borra el usuario del mapa
		return
	}
	return errNoSesion //Si no existe la conexión se manda el errNoSesion
}

//sendCommand envía el comando al servidor remoto
func sendCommand(m Messaging) (result string, err error) {
	usr := mapaUsuarios[m.Sender.ID]
	if usr.conn == nil { //Si no hay conexión se regresa el errNoSesion
		err = errNoSesion
		return
	}
	log.Println(m.Sender.ID, "Comando a enviar:", m.Message.Text)
	session, _ := usr.conn.NewSession()                 //Se crea una sesión para mandar comandos
	data, err := session.CombinedOutput(m.Message.Text) //Ejecuta el comando y regresa la salida estandar y el error
	result = string(data)                               //Se convierten los bytes a string
	session.Close()                                     //Se cierra la sesión
	return
}

//allowIP verifica que se contenga el ip dentro del arreglo de ips
func allowIP(ip string, ips []string) bool {
	for i := range ips {
		if ips[i] == ip { //Cuando se encuentra el IP en el arreglo regresa verdadero
			return true
		}
	}
	return false //Si no se encuentra regresa falso
}

//publicKeyFile es el metodo encargado de conseguir la autenticación SSH utilizando las llaves
func publicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file) //Se lee la llave
	checkErr(err)
	key, err := ssh.ParsePrivateKey(buffer) //Se convierte la llave privada
	checkErr(err)
	return ssh.PublicKeys(key) //Regresa la autenticación
}

//checkErr es utilizado para parar el servidor en caso que existe un error fatal
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

//updatePath guarda la ruta del directorio actual
func updatePath(userid, path string) {
	if mapaUsuarios[userid].conn != nil {
		u := mapaUsuarios[userid]
		u.lastPath = path
		mapaUsuarios[userid] = u
		fmt.Println("Guardardo path ", mapaUsuarios[userid])
	}
}

//getPath obtiene la ruta actual
func getPath(userid string) (path string) {
	if mapaUsuarios[userid].conn != nil {
		path = mapaUsuarios[userid].lastPath
		fmt.Println("Obteniendo path", path)
	}
	return
}
