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

var addressLocal string
var hostLocales []string

var players_local []string
var teams_local []string

var players []Player
var teams []Team
var winningTeamList []int

var num_int int

func main() {
	//Lectura por consola del host origin
	brInput := bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese el puerto del host local iniciador: ")
	puertoLocal, _ := brInput.ReadString('\n')
	puertoLocal = strings.TrimSpace(puertoLocal)

	addressLocal = fmt.Sprintf("localhost:%s", puertoLocal)

	//Cuantos equipos jugaran
	brInput = bufio.NewReader(os.Stdin)
	fmt.Printf("Ingrese la cantidad de equipos a jugar: ")
	num, _ := brInput.ReadString('\n')
	num = strings.TrimSpace(num)
	num_int, _ = strconv.Atoi(num)

	// Habilitar escuchar
	ln, _ := net.Listen("tcp", addressLocal)
	defer ln.Close()
	for {
		conn, _ := ln.Accept()
		go recibirData(conn)
	}

}

func enviar(host string, msg string) {
	conn, _ := net.Dial("tcp", host)
	defer conn.Close()
	fmt.Fprintf(conn, "%s\n", msg)
}

func recibirData(conn net.Conn) {
	//Recibo host local el que envia y guardo en el arreglo
	defer conn.Close()
	reader := bufio.NewReader(conn)
	//Leer el host string enviado
	host_string, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Se cerro la conexion dado que no se pudo leer el dato")
		conn.Close()
	} else {
		hostLocales = append(hostLocales, strings.TrimSpace(host_string))
		fmt.Printf("Se esta recibiendo un nuevo Host: %s\n", hostLocales)
	}

	// Leer el json del jugador
	jsonplayer, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Se cerro la conexion dado que no se pudo leer el dato de json player")
		conn.Close()
	} else {
		jsonplayer = strings.TrimSpace(jsonplayer)
		players_local = append(players_local, jsonplayer)
		fmt.Printf("Se esta recibiendo una nueva lista de jugadores %s\n", jsonplayer)
	}
	// Leer el json del equipo
	jsonteam, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Se cerro la conexion dado que no se pudo leer el dato de json team")
		conn.Close()
	} else {
		jsonteam = strings.TrimSpace(jsonteam)
		teams_local = append(teams_local, jsonteam)
		fmt.Printf("Se esta recibiendo un nuevo equipo %s\n", jsonteam)
	}
	fmt.Printf("El total de equipos es: %d\n", len(hostLocales))
	fmt.Println("----------------------------------------------------------------")
	if len(hostLocales) == num_int {
		juego()
	}
}

func juego() {
	// Definir variables para almacenar las listas
	fmt.Println("Ingresamos al juego felicidades")
	var listaJSON_players []Player
	var listaJSON_teams []Team

	// Convertir la cadena de texto en una lista de objetos JSON
	for _, val := range players_local {
		var playerJSON []Player
		err := json.Unmarshal([]byte(val), &playerJSON)
		if err != nil {
			fmt.Println("Error al convertir la cadena de texto en lista JSON:", err)
			return
		}
		listaJSON_players = append(listaJSON_players, playerJSON...)
	}

	// Imprimir el nuevo JSON
	fmt.Println("Nuevo JSON con la lista de players")
	nuevoJSON_players, err := json.Marshal(listaJSON_players)
	if err != nil {
		fmt.Println("Error al crear el nuevo JSON:", err)
		return
	}
	fmt.Println(string(nuevoJSON_players))

	// Convertir la cadena de texto en una lista de objetos JSON
	for _, val := range teams_local {
		var teamJSON []Team
		err := json.Unmarshal([]byte(val), &teamJSON)
		if err != nil {
			fmt.Println("Error al convertir la cadena de texto en lista JSON:", err)
			return
		}
		listaJSON_teams = append(listaJSON_teams, teamJSON...)
	}

	// Imprimir el nuevo JSON
	fmt.Println("Nuevo JSON con la lista de teams")
	nuevoJSON_teams, err := json.Marshal(listaJSON_teams)
	if err != nil {
		fmt.Println("Error al crear el nuevo JSON:", err)
		return
	}
	fmt.Println(string(nuevoJSON_teams))

	json.Unmarshal([]byte(string(nuevoJSON_players)), &players)
	json.Unmarshal([]byte(string(nuevoJSON_teams)), &teams)
	fmt.Printf("Ingrese desarializar y comenzar juego\n")
	fmt.Println("----------------------------------------------------------------")
	fmt.Println("----------------------------------------------------------------")
	gameOver := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(len(players)) // tienes que realizar si o si 20 operacion // 16 operaciones  4 veces
	for i := 0; i < len(players); i++ {
		go func(player *Player) {
			defer wg.Done()
			for {
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
					winningTeamList = append(winningTeamList, player.Teamint)
					return
				}
			}
		}(&players[i])
	}
	go func() {
		wg.Wait()
		close(gameOver)
	}()

	select {
	case <-gameOver:
		winningTeam := getWinningTeam(teams)
		fmt.Printf("¡El equipo %d ha ganado con %d puntos!\n", winningTeam.IDmint, winningTeam.Points)
		for i := 0; i < len(hostLocales); i++ {
			if i == winningTeam.IDmint {
				fmt.Printf("%s\n", hostLocales[i])
				enviar(hostLocales[i], "Equipo ganador")
			} else {
				fmt.Printf("%s\n", hostLocales[i])
				enviar(hostLocales[i], "Tu equipo perdio")
			}
		}
	}

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

	estado_Actual := fmt.Sprintf("Jugador %d (equipo %d) juega %s . Jugador %d (equipo %d) juega %s.\n",
		player1.ID, player1.Teamint, handSigns[hand1],
		player2.ID, player2.Teamint, handSigns[hand2])

	enviar(hostLocales[player1.Teamint], estado_Actual)
	enviar(hostLocales[player2.Teamint], estado_Actual)

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
		if team.IDmint == winningTeamList[0] {
			return team
		}

		//if team.Points >= team.PointsTarget {
		//	return team
		//}
	}
	return Team{}
}
