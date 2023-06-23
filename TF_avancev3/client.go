package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var addressRemoto string
var addressLocal string

type Player struct {
	ID       int
	Team     string
	Teamint  int
	Position int
	Meta     bool
	Mutex    sync.Mutex
}

type Team struct {
	ID           string
	IDmint       int
	Points       int
	PointsTarget int
	Mutex        sync.Mutex
}
type Host struct {
	HOST string
}

var players []Player

var teams []Team
var numTeam int

var upgrader = websocket.Upgrader{}

func main() {
	// Lectura por consola del host origin
	fmt.Printf("Soy el host local %s\n", myIp())
	addressLocal = fmt.Sprintf("%s:8000", myIp())

	// Lectura por consola del host destino
	brInput := bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese la ip remota: ")
	ipRemoto, _ := brInput.ReadString('\n')
	ipRemoto = strings.TrimSpace(ipRemoto)
	addressRemoto = fmt.Sprintf("%s:8000", ipRemoto)

	// Lectura por consola del equipo
	brInput = bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese el numero del equipo: ")
	num, _ := brInput.ReadString('\n')
	num = strings.TrimSpace(num)
	numTeam, _ = strconv.Atoi(num)

	enviar()

	http.HandleFunc("/ws", handleWebSocket)
	go http.ListenAndServe(addressLocal, nil)

	select {}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error al actualizar la conexión:", err)
		return
	}
	defer conn.Close()

	manejador(conn)
}

func enviar() {
	conn, _, err := websocket.DefaultDialer.Dial("ws://"+addressRemoto+"/", nil)
	if err != nil {
		log.Fatal("Error al establecer la conexión:", err)
	}
	defer conn.Close()

	numTeams := 1
	numPlayersPerTeam := 4

	teams = make([]Team, numTeams)
	for i := 0; i < numTeams; i++ {
		teams[i] = Team{ID: addressLocal, PointsTarget: numPlayersPerTeam, IDmint: numTeam}
	}

	players = make([]Player, numTeams*numPlayersPerTeam)
	for i := 0; i < numTeams; i++ {
		for j := 0; j < numPlayersPerTeam; j++ {
			playerID := i*numPlayersPerTeam + j
			players[playerID] = Player{
				ID:       playerID,
				Team:     addressLocal,
				Position: 0,
				Meta:     false,
				Teamint:  numTeam,
			}
		}
	}
	arryJsonPlayer, _ := json.Marshal(players)
	jsonStringPlayer := string(arryJsonPlayer)

	arryJsonTeam, _ := json.Marshal(teams)
	jsonStringTeam := string(arryJsonTeam)

	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s\n%s\n%s\n", addressLocal, jsonStringPlayer, jsonStringTeam)))
}

func manejador(conn *websocket.Conn) {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error al leer el mensaje:", err)
			return
		}
		fmt.Printf("%s\n", strings.TrimSpace(string(msg)))
	}
}

func myIp() string {
	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		if strings.HasPrefix(iface.Name, "eth0") {
			addrs, _ := iface.Addrs()
			for _, addr := range addrs {
				switch v := addr.(type) {
				case *net.IPNet:
					return v.IP.String()
				case *net.IPAddr:
					return v.IP.String()
				}
			}
		}
	}
	return ""
}
