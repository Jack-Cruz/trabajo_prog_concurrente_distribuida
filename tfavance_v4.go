package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
)

// Bitacora de direcciones de miembros de la red
var bitacora_red []string

// Dirrecion de red del nodo
var dir_nodo string

//Pueros del servicio en el nodo

const (
	puerto_registro  = 8000 // o puerto de escucha
	puerto_notifica  = 8001
	puerto_proceso   = 8002 //Servicio (Enviar, recibir)
	puerto_solicitud = 8003
)

type Player struct {
	ID       int
	Team     int
	Teamint  int
	Position int
	Meta     bool
}

type Team struct {
	ID           int
	IDmint       int
	Points       int
	PointsTarget int
}
type Mensaje struct {
	NumTeams    int
	IndexSearch int
}

var players_local []string
var teams_local []string
var playersAlonesString []string

var players []Player
var teams []Team

var playersAlones []Player

var players_alone []Player
var teams_alone []Team

var winningTeamList []int

var destino_meta = 2

var numTeams = 0
var IndexSearch = 0
var mutex sync.Mutex

func main() {
	//Identificarse
	dir_nodo = obtenerIP()
	fmt.Printf("Hola soy %s\n", dir_nodo)

	//Rol de servidor (Escucha) para registrar nuevos nodos
	go registrarServidor()

	//Servicio del nodo (Escucha) para el Hot Potato
	go registrarServicioHP()

	//Rol Cliente
	// Solicitar unirse a la red
	br := bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese la IP del nodo  remoto a conectarse:")
	remotehost, _ := br.ReadString('\n')
	remotehost = strings.TrimSpace(remotehost)

	//Para que el nodo se una a la red existente
	//Este sera un cliente
	if remotehost != "" {
		registrarCliente(remotehost) //Quiero conectarme a tu red
	}
	//Rol de servidor
	escucharNotificaciones()
}

func obtenerIP() string {
	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		if strings.HasPrefix(iface.Name, "Ethernet") {
			addrs, _ := iface.Addrs()
			for _, addr := range addrs {
				switch v := addr.(type) {
				//case *net.IPNet:
				//	return v.IP.String()
				//case *net.IPAddr:
				//	return v.IP.String()
				case *net.IPNet:
					if v.IP.To4() != nil {
						return v.IP.To4().String()
					}
				}
			}
		}
	}
	return "127.0.0.1"
}
func registrarServidor() {
	//Modo Escucha siempre
	dirNodo := fmt.Sprintf("%s:%d", dir_nodo, puerto_registro)
	listener, _ := net.Listen("tcp", dirNodo)
	defer listener.Close()
	//Estar activo siempre aceptando las conexiones
	for {
		//aceptar las conexiones
		conn, _ := listener.Accept()
		go manejadorRegistro(conn)
	}
}

func manejadorRegistro(conn net.Conn) {
	defer conn.Close()

	//Recibir la llamda de nuevo host, llega la IP del host
	br := bufio.NewReader(conn)
	remoteIp, _ := br.ReadString('\n')
	remoteIp = strings.TrimSpace(remoteIp)

	//Devolver al nuevo nodo la bitacora que este nodo guarda, el que ya estaba guardando
	bytesBitacora, _ := json.Marshal(bitacora_red)
	fmt.Fprintf(conn, "%s\n", string(bytesBitacora))
	//Notificamos al resto de los nodos que llego una nueva conexion

	notificarTodos(remoteIp)

	//Actualizar su bitacora
	bitacora_red = append(bitacora_red, remoteIp)

	println("Bitacora actualizada")
}
func notificarTodos(remoteIp string) {

	for _, dir := range bitacora_red {
		notificar(dir, remoteIp)
	}
}
func notificar(dir, ip_remota string) {
	//Comunicar
	host_remoto := fmt.Sprintf("%s:%d", dir, puerto_notifica)
	conn, _ := net.Dial("tcp", host_remoto)
	defer conn.Close()
	fmt.Fprintf(conn, "%s\n%s\n", "IP", ip_remota)
}
func registrarServicioHP() {
	//Modo Escucha siempre
	dirNodo := fmt.Sprintf("%s:%d", dir_nodo, puerto_proceso)
	listener, _ := net.Listen("tcp", dirNodo)
	defer listener.Close()
	//Estar activo siempre aceptando las conexiones
	for {
		//aceptar las conexiones
		conn, _ := listener.Accept()
		go manejadorProcesoHP(conn)
	}
}
func manejadorProcesoHP(conn net.Conn) {
	defer conn.Close()
	br := bufio.NewReader(conn)
	msg, _ := br.ReadString('\n')
	msg = strings.TrimSpace(msg)
	fmt.Printf("Recibiendo a %s\n", msg)
	switch msg {
	case "All":
		msgPlayers, _ := br.ReadString('\n')
		msgPlayers = strings.TrimSpace(msgPlayers)
		fmt.Printf("Recibiendo a %s\n", msgPlayers)
		//Agregando a un arrray la lista de jugadores que llegan del inicializador y que llegaron de otros nodos
		players_local = append(players_local, msgPlayers)
		//Conviertiendo el array string a un arrat Players
		for _, player := range players_local {
			var playerString []Player
			err := json.Unmarshal([]byte(player), &playerString)
			if err != nil {
				fmt.Println("Error al convertir la cadena de texto en lista JSON:", err)
				return
			}
			players = append(players, playerString...)
		}
		for _, player := range players {
			if player.ID == IndexSearch {
				for i := 0; i < numTeams; i++ {
					if player.Team == i {
						players_alone = append(players_alone, player)
					}
				}
			}
		}
		fmt.Println("Impresion de prueba de players_alone")
		fmt.Println(players_alone)
		enviarPlayers("Alones", players_alone)
	case "NewPlayers":
		for _, player := range players {
			if player.ID == IndexSearch {
				for i := 0; i < numTeams; i++ {
					if player.Team == i {
						players_alone = append(players_alone, player)
					}
				}
			}
		}
		enviarPlayers("Alones", players_alone)
	case "Alones":
		players_alone = []Player{}
		playerAlones, _ := br.ReadString('\n')
		playerAlones = strings.TrimSpace(playerAlones)
		fmt.Printf("Recibiendo Alones: %s\n", playerAlones)
		playersAlonesString = append(playersAlonesString, playerAlones)
		//Convirtiendo los players Alones
		for _, player := range playersAlonesString {
			var playerString []Player
			err := json.Unmarshal([]byte(player), &playerString)
			if err != nil {
				fmt.Println("Error al convertir la cadena de texto en lista JSON:", err)
				return
			}
			players_alone = append(playersAlones, playerString...)
		}
		for _, player := range players_alone {
			player.Position++
			if destino_meta != player.Position {
				enviarPlayer("Alone", player)
			} else {
				player.Meta = true
				fmt.Println("Llego a la meta:")
				fmt.Println(player)
			}
		}

	case "Alone":
		playerString, _ := br.ReadString('\n')
		playerString = strings.TrimSpace(playerString)
		var playerOficial Player
		var cont int = 0
		json.Unmarshal([]byte(playerString), &playerOficial)
		for _, player := range players_alone {
			player.Position++
			if playerOficial.Team != player.Team && playerOficial.Position == player.Position && playerOficial.Meta == player.Meta {
				playRockPaperScissors(&playerOficial, &player)
			}
			if playerOficial.ID == player.ID && playerOficial.Team == player.Team {
				player.Position = playerOficial.Position
				enviarPlayer("Alone", player)
			}
			if destino_meta == player.Position {
				player.Meta = true
				fmt.Println("Llego a la meta alone: ")
				fmt.Println(player)
			} else {
				enviarPlayer("Alone", player)
			}
			if player.Meta {
				cont++
			}
		}
		if cont == 4 {
			indice := IndexSearch
			indice++
			IndexSearch = indice % 4
			enviarNewPlayers("NewPlayers")
		}

	}

}
func enviarProximoNodo(num int) {
	//Seleccionar de forma aleatoria el proximo nodo
	indice := rand.Intn(len(bitacora_red))
	//Enviando mesnaje a
	fmt.Printf("Enviando el numero %d a %s\n", num, bitacora_red[indice])
	nextNodeIP := fmt.Sprintf("%s:%d", bitacora_red[indice], puerto_proceso)
	conn, _ := net.Dial("tcp", nextNodeIP)
	defer conn.Close()
	//Enviar el numero al siguiente nodo
	fmt.Fprintln(conn, num)
}

func enviarNextNode(player []Player, team []Team) {
	indice := rand.Intn(len(bitacora_red))
	fmt.Printf("Enviando el estado de juego  a %s\n", bitacora_red[indice])
	nextNodeIP := fmt.Sprintf("%s:%d", bitacora_red[indice], puerto_proceso)
	conn, _ := net.Dial("tcp", nextNodeIP)
	defer conn.Close()
	arryJsonPlayer, _ := json.Marshal(player)
	jsonStringPlayer := string(arryJsonPlayer)

	arryJsonTeam, _ := json.Marshal(team)
	jsonStringTeam := string(arryJsonTeam)
	fmt.Fprintf(conn, "%s\n%s\n", jsonStringPlayer, jsonStringTeam)
}

func enviarPlayer(tipo string, player Player) {
	indice := rand.Intn(len(bitacora_red))
	fmt.Printf("Enviando el estado de juego Alone a %s\n", bitacora_red[indice])
	nextNodeIP := fmt.Sprintf("%s:%d", bitacora_red[indice], puerto_proceso)
	conn, _ := net.Dial("tcp", nextNodeIP)
	playerAlonejson, _ := json.Marshal(player)
	fmt.Println(string(playerAlonejson))
	fmt.Fprintf(conn, "%s\n%s\n", tipo, string(playerAlonejson))
	defer conn.Close()
}
func enviarPlayers(tipo string, player []Player) {
	indice := rand.Intn(len(bitacora_red))
	fmt.Printf("Enviando el estado de juego Alones  a %s\n", bitacora_red[indice])
	nextNodeIP := fmt.Sprintf("%s:%d", bitacora_red[indice], puerto_proceso)
	conn, _ := net.Dial("tcp", nextNodeIP)
	defer conn.Close()
	playerAlonejson, _ := json.Marshal(player)
	fmt.Println(string(playerAlonejson))
	fmt.Fprintf(conn, "%s\n%s\n", tipo, string(playerAlonejson))
}
func enviarNewPlayers(tipo string) {
	indice := rand.Intn(len(bitacora_red))
	fmt.Printf("Enviando el estado de juego NewPlayers  a %s\n", bitacora_red[indice])
	nextNodeIP := fmt.Sprintf("%s:%d", bitacora_red[indice], puerto_proceso)
	conn, _ := net.Dial("tcp", nextNodeIP)
	defer conn.Close()
	fmt.Fprintf(conn, "%s\n", tipo)
}

func playRockPaperScissors(player1 *Player, player2 *Player) {
	handSigns := [3]string{"Piedra", "Papel", "Tijeras"}

	hand1 := rand.Intn(3)
	hand2 := rand.Intn(3)

	fmt.Printf("Jugador %d (equipo %d) juega %s . Jugador %d (equipo %d) juega %s.\n",
		player1.ID, player1.Team, handSigns[hand1],
		player2.ID, player2.Team, handSigns[hand2])

	if hand1 == hand2 {
		return
	}

	switch hand1 {
	case 0: // Piedra
		if hand2 == 1 { // Papel
			player2.Position++
			player1.Position = 0
		} else { // Tijeras
			player1.Position++
			player2.Position = 0

		}
	case 1: // Papel
		if hand2 == 0 { // Piedra
			player1.Position++
			player2.Position = 0
		} else { // Tijeras
			player2.Position++
			player1.Position = 0

		}
	case 2: // Tijeras
		if hand2 == 0 { // Piedra
			player2.Position++
			player1.Position = 0
		} else { // Papel
			player1.Position++
			player2.Position = 0
		}
	}

}

func registrarCliente(remotehost string) {
	//Se usa cuando ya existe una red y se quiere unir a ella
	remote_Host := fmt.Sprintf("%s:%d", remotehost, puerto_registro)
	conn, _ := net.Dial("tcp", remote_Host)
	defer conn.Close()

	//Enviamos la Ip de este nodo al host remoto( servidor, entiendo)
	// DIR NODO guarda la direccion IP del nodo actual
	fmt.Fprintf(conn, "%s\n", dir_nodo)

	//Como respuesta el host remoto envia su bitacora
	br := bufio.NewReader(conn)
	strbitacora, _ := br.ReadString('\n')
	//Decodificar la bitacora
	var dirTemp []string
	json.Unmarshal([]byte(strbitacora), &dirTemp)
	//Se actualiza la bitacora de direciones del nodo
	bitacora_red = append(dirTemp, remotehost)
	fmt.Println("Bitacora actualizada")
	fmt.Println(bitacora_red)
}

func escucharNotificaciones() {
	//Modo escucha
	dirNode := fmt.Sprintf("%s:%d", dir_nodo, puerto_notifica)
	listener, _ := net.Listen("tcp", dirNode)
	defer listener.Close()
	for {
		//Aceptar conexiones
		conn, _ := listener.Accept()
		go manejadorNotificacion(conn)
	}
}

func manejadorNotificacion(conn net.Conn) {
	defer conn.Close()

	//Recibir las notificiaciones, que le llega el IP del nodo conectandose
	br := bufio.NewReader(conn)
	strTipo, _ := br.ReadString('\n')
	strTipo = strings.TrimSpace(strTipo)
	switch strTipo {
	case "Informacion":
		strMensaje, _ := br.ReadString('\n')
		strMensaje = strings.TrimSpace(strMensaje)
		var mensaje Mensaje
		json.Unmarshal([]byte(strMensaje), &mensaje)
		fmt.Printf("Manejando la informacion: %s\n", strMensaje)
		numTeams = mensaje.NumTeams
		IndexSearch = mensaje.IndexSearch
		fmt.Println(numTeams)
		fmt.Println(IndexSearch)

	case "IP":
		strIP, _ := br.ReadString('\n')
		strIP = strings.TrimSpace(strIP)
		//Cada nodo va a registrar la IP en su bitacora
		bitacora_red = append(bitacora_red, strIP)
		fmt.Println("Cliente notificado:")
		fmt.Println(bitacora_red)
	}

}
