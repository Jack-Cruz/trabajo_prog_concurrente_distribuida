package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
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

//var player_anterior Player
//var team_anterior Team
//var player_actual Player
//var team_actual Team

var players []Player

var teams []Team

var cont int          //Contar el numero de equipos
var contChan chan int //valor que nos indica que habra 3 equipos
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

	ln, _ := net.Listen("tcp", addressLocal)
	defer ln.Close()

	contChan = make(chan int, 1)
	contChan <- 0

	for {
		conn, _ := ln.Accept()
		go manejador(conn)
	}
}

func manejador(conn net.Conn) {
	num, jsonplayer, jsonteam, err := recibir(conn)
	fmt.Printf("Contador actual %d\n", cont)
	// fmt.Printf("num %d, json %s, team %s\n\n", num, jsonplayer, jsonteam)
	fmt.Printf("num %d, json %s\n\n", num, jsonplayer)
	if err != nil {
		fmt.Println("Error al recibir los datos de inicio:", err)
		return
	}
	cont = <-contChan
	if num == 0 && len(jsonplayer) == 0 && len(jsonteam) == 0 {
		numTeams := 1
		numPlayersPerTeam := 1

		temas := make([]Team, numTeams)
		for i := 0; i < numTeams; i++ {
			temas[i] = Team{ID: conn.LocalAddr().String(), PointsTarget: numPlayersPerTeam, IDmint: numTeam}
		}

		jugadores := make([]Player, numTeams*numPlayersPerTeam)
		for i := 0; i < numTeams; i++ {
			for j := 0; j < numPlayersPerTeam; j++ {
				playerID := i*numPlayersPerTeam + j
				jugadores[playerID] = Player{
					ID:       playerID,
					Team:     conn.LocalAddr().String(),
					Position: 0,
					Meta:     false,
					Teamint:  numTeam,
				}
			}
		}
		cont++
		arryJsonPlayer, _ := json.Marshal(jugadores)
		jsonStringPlayer := string(arryJsonPlayer)

		arryJsonTeam, _ := json.Marshal(temas)
		jsonStringTeam := string(arryJsonTeam)
		contChan <- cont
		enviarTeam(0, jsonStringPlayer, jsonStringTeam)

	}
	if len(jsonplayer) > 0 && len(jsonteam) > 0 && cont < 3 {
		fmt.Println("llego a la segunda fase")
		numTeams := 1
		numPlayersPerTeam := 1

		temas := make([]Team, numTeams)
		for i := 0; i < numTeams; i++ {
			temas[i] = Team{ID: conn.LocalAddr().String(), PointsTarget: numPlayersPerTeam, IDmint: numTeam}
		}

		jugadores := make([]Player, numTeams*numPlayersPerTeam)
		for i := 0; i < numTeams; i++ {
			for j := 0; j < numPlayersPerTeam; j++ {
				playerID := i*numPlayersPerTeam + j
				jugadores[playerID] = Player{
					ID:       playerID,
					Team:     conn.LocalAddr().String(),
					Position: 0,
					Meta:     false,
					Teamint:  numTeam,
				}
			}
		}
		cont++
		arryJsonPlayer, _ := json.Marshal(jugadores)
		jsonStringPlayer := string(arryJsonPlayer)

		arryJsonTeam, _ := json.Marshal(temas)
		jsonStringTeam := string(arryJsonTeam)

		var player_anterior []Player
		var team_anterior []Team

		json.Unmarshal([]byte(jsonplayer), &player_anterior)
		json.Unmarshal([]byte(jsonteam), &team_anterior)

		var player_actual []Player
		var team_actual []Team

		json.Unmarshal([]byte(jsonStringPlayer), &player_actual)
		json.Unmarshal([]byte(jsonStringTeam), &team_actual)

		// Combinar datos
		player_anterior = append(player_anterior, player_actual...)
		team_anterior = append(team_anterior, team_actual...)

		// Codificar el objeto combinado a JSON
		jsonCombPlayer, _ := json.Marshal(player_anterior)
		jsonCombTeam, _ := json.Marshal(team_anterior)

		//fmt.Printf("En segunda fase %s, team %s\n", string(jsonCombPlayer), string(jsonCombTeam))
		contChan <- cont
		enviarTeam(0, string(jsonCombPlayer), string(jsonCombTeam))
	}
	if cont >= 3 { //start game and pass all teams information}
		fmt.Println("llego a la tercera fase")
	return	
	//fmt.Printf("Ingrese sin deserializar %s , %s \n", jsonplayer, jsonteam)
		json.Unmarshal([]byte(jsonplayer), &players)
		json.Unmarshal([]byte(jsonteam), &teams)
		fmt.Printf("Ingrese desarializar y comenzar juego\n")
		gameOver := make(chan struct{})
		wg := sync.WaitGroup{}
		wg.Add(len(players))
		
		for i := 0; i < len(players); i++ {
			go func(player *Player) {
				defer wg.Done()
				for {
					fmt.Printf("Estoy dentro del juego \n")
					player.Mutex.Lock()
					player.Position++
					player.Mutex.Unlock()

					time.Sleep(time.Millisecond * time.Duration(rand.Intn(500)))

					opponent := getOpponent(players, player)
					if opponent != nil && opponent.Position == player.Position && opponent.Meta == player.Meta {
						playRockPaperScissors(player, opponent)
					}

					player.Mutex.Lock()
					if player.Position == 20 {
						teams[player.Teamint].Mutex.Lock()
						teams[player.Teamint].Points += 1
						teams[player.Teamint].Mutex.Unlock()
						player.Meta = true
						fmt.Printf("Jugador %d (equipo %d) se ha retirado y su equipo tiene %d.\n",
							player.ID, player.Teamint, teams[player.Teamint].Points)
					}
					player.Mutex.Unlock()

					if teams[player.Teamint].Points >= teams[player.Teamint].PointsTarget {
						close(gameOver)
						return
					}
				}
			}(&players[i])
		}
		select {
		case <-gameOver:
			winningTeam := getWinningTeam(teams)
			fmt.Printf("¡El equipo %d ha ganado con %d puntos!\n", winningTeam.IDmint, winningTeam.Points)
		}
	}
	defer conn.Close()
}
func getOpponent(players []Player, player *Player) *Player {
	for _, p := range players {
		if p.Team != player.Team && p.Position == player.Position {
			return &p
		}
	}
	return nil
}

func playRockPaperScissors(player1 *Player, player2 *Player) {
	handSigns := [3]string{"Piedra", "Papel", "Tijeras"}

	hand1 := rand.Intn(3)
	hand2 := rand.Intn(3)

	fmt.Printf("Jugador %d (equipo %d) juega %s . Jugador %d (equipo %d) juega %s.\n",
		player1.ID, player1.Teamint, handSigns[hand1],
		player2.ID, player2.Teamint, handSigns[hand2])

	if hand1 == hand2 {
		return
	}

	player1.Mutex.Lock()
	player2.Mutex.Lock()

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

	player2.Mutex.Unlock()
	player1.Mutex.Unlock()
}

func getWinningTeam(teams []Team) Team {
	for _, team := range teams {
		if team.Points >= team.PointsTarget {
			return team
		}
	}
	return Team{}
}

func recibir(conn net.Conn) (num int, jsonplayer string, jsonteam string, err error) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	// Leer el número
	numStr, err := reader.ReadString('\n')
	if err != nil {
		return 0, "", "", err
	}
	numStr = strings.TrimSpace(numStr)
	num, err = strconv.Atoi(numStr)
	if err != nil {
		return 0, "", "", err
	}

	// Leer el json del jugador
	jsonplayer, err = reader.ReadString('\n')
	if err != nil {
		return 0, "", "", err
	}
	jsonplayer = strings.TrimSpace(jsonplayer)

	// Leer el json del equipo
	jsonteam, err = reader.ReadString('\n')
	if err != nil {
		return 0, "", "", err
	}
	jsonteam = strings.TrimSpace(jsonteam)

	return num, jsonplayer, jsonteam, nil
}

func enviarTeam(num int, jsonPlayer string, jsonTeam string) {
	conn, _ := net.Dial("tcp", addressRemoto)
	defer conn.Close()
	//fmt.Printf("Dentro de enviar %s , team %s\n", jsonPlayer, jsonTeam)
	fmt.Fprintf(conn, "%d\n%s\n%s\n\n", num, jsonPlayer, jsonTeam)
}
