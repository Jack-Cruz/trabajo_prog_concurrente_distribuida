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
	fmt.Printf("Soy el host local %s\n", myIp())
	addressLocal = fmt.Sprintf("%s:8000", myIp())

	//Lectura por consola del host destino
	brInput := bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese la ip remota: ")
	ipRemoto, _ := brInput.ReadString('\n')
	ipRemoto = strings.TrimSpace(ipRemoto)
	addressRemoto = fmt.Sprintf("%s:8000", ipRemoto)

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

func myIp() string { // mandrakeando ando
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