package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
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

var players []Player

var teams []Team
var numTeam int

func main() {
	//Lectura por consola del host origin
	brInput := bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese el puerto del host local: ")
	puertoLocal, _ := brInput.ReadString('\n')
	puertoLocal = strings.TrimSpace(puertoLocal)

	addressLocal = fmt.Sprintf("localhost:%s", puertoLocal)

	//Lectura por consola del host destino
	brInput = bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese el puerto del host remoto: ")
	puertoRemoto, _ := brInput.ReadString('\n')
	puertoRemoto = strings.TrimSpace(puertoRemoto)
	addressRemoto = fmt.Sprintf("localhost:%s", puertoRemoto)

	//Lectura por consola del equipo
	brInput = bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese el numero del equipo: ")
	num, _ := brInput.ReadString('\n')
	num = strings.TrimSpace(num)
	numTeam, _ = strconv.Atoi(num)

	enviar()
	ln, _ := net.Listen("tcp", addressLocal)
	defer ln.Close()

	for {
		conn, _ := ln.Accept()
		go manejador(conn)
	}
}
func enviar() {
	conn, _ := net.Dial("tcp", addressRemoto)
	defer conn.Close()

	numTeams := 1
	numPlayersPerTeam := 4

	teams := make([]Team, numTeams)
	for i := 0; i < numTeams; i++ {
		teams[i] = Team{ID: addressLocal, PointsTarget: numPlayersPerTeam, IDmint: numTeam}
	}

	players := make([]Player, numTeams*numPlayersPerTeam)
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

	fmt.Fprintf(conn, "%s\n%s\n%s\n", addressLocal, jsonStringPlayer, jsonStringTeam)
}

func manejador(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	msg, _ := reader.ReadString('\n')
	fmt.Printf("%s\n", strings.TrimSpace(msg))
}
