package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
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

type HostMessage struct {
	Host string `json:"host"`
}

type PlayerMessage struct {
	Players string `json:"players"`
}

type TeamMessage struct {
	Teams string `json:"teams"`
}

var addressLocal string
var hostLocales []string

var players_local []string
var teams_local []string

var players []Player
var teams []Team
var winningTeamList []int

var num_int int
var gameOver chan struct{}

var clients []websocket.Conn

func main() {
	// Lectura por consola del host origin
	fmt.Printf("La ip host local iniciador: %s\n", myIP())
	addressLocal = fmt.Sprintf("%s:8000", myIP())

	// Cuantos equipos jugaran
	brInput := bufio.NewReader(os.Stdin)
	fmt.Printf("Ingrese la cantidad de equipos a jugar: ")
	num, _ := brInput.ReadString('\n')
	num = strings.TrimSpace(num)
	num_int, _ = strconv.Atoi(num)

	http.HandleFunc("/", handleWebSocket)
	http.HandleFunc("/ws", wsEndpoint)
	http.HandleFunc("/game", serveGameHTML)
	go http.ListenAndServe(addressLocal, nil)

	select {}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	// fmt.Fprintf(w, "Hello World")
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	// upgrade this connection to a WebSocket
	// connection

	ws, _ := upgrader.Upgrade(w, r, nil)
	clients = append(clients, *ws)
	for {
		// READ MESSAGE FROM BROWSER
		msgType, msg, err := ws.ReadMessage()
		if err != nil {
			return
		}

		// PRINT MESSAGE IN YOU CONSOLE TERMINAL
		fmt.Printf("%s send: %s\n", ws.RemoteAddr(), string(msg))

		// LOOP IF MESSAGE FOUND AND SEND AGAIN TO CLIENT FOR
		// WRITE IN YOU BROWSER
		for _, client := range clients {
			if err = client.WriteMessage(msgType, msg); err != nil {
				return
			}
		}

	}
	//reader(ws)
}
func serveGameHTML(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "game.html")
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if err != nil {
		log.Println("Error al actualizar la conexión WebSocket:", err)
		return
	}
	defer conn.Close()
	clients = append(clients, *conn)

	// Recibo host local el que envia y guardo en el arreglo
	_, messageBytes, err := conn.ReadMessage()
	if err != nil {
		log.Println("Error al leer el mensaje del cliente ", err)
		return
	}
	newHost := string(messageBytes)
	data := strings.Split(newHost, "\n")
	hostLocales = append(hostLocales, data[0])
	fmt.Printf("Se ha unido un nuevo host a la lista %s\n ", hostLocales)
	//Recibo jsonPlayers
	players_local = append(players_local, data[1])
	fmt.Printf("Se esta recibiendo una nueva lista de jugadores %s\n", data[1])
	//Recibo jsonTeam
	teams_local = append(teams_local, data[2])
	fmt.Printf("Se esta recibiendo un nuevo equipo %s\n", data[2])

	for _, client := range clients {
		if err = client.WriteMessage(1, []byte(newHost)); err != nil {
			return
		}
		fmt.Printf("El total de equipos es: %d\n", len(hostLocales))
		fmt.Println("----------------------------------------------------------------")
		if len(hostLocales) == num_int {
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
			fmt.Printf("Ingrese deserializar y comenzar juego\n")
			fmt.Println("----------------------------------------------------------------")
			fmt.Println("----------------------------------------------------------------")
			gameOver = make(chan struct{})
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
						if player.Position == 15 {
							teams[player.Teamint].Mutex.Lock()
							teams[player.Teamint].Points += 1
							teams[player.Teamint].Mutex.Unlock()
							player.Meta = true
							fmt.Printf("Jugador %d (equipo %d) se ha retirado y su equipo tiene %d.\n",
								player.ID, player.Teamint, teams[player.Teamint].Points)
							message := fmt.Sprintf("Jugador %d (equipo %d) se ha retirado y su equipo tiene %d.\n",
								player.ID, player.Teamint, teams[player.Teamint].Points)
							clients[0].WriteMessage(1, []byte(message))
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

	}
}

func enviar(host string, msg string) {
	conn, _, err := websocket.DefaultDialer.Dial("ws://"+host+"/ws", nil)
	if err != nil {
		log.Println("Error al establecer la conexión WebSocket:", err)
		return
	}
	defer conn.Close()

	err = conn.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		log.Println("Error al enviar el mensaje:", err)
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
	fmt.Printf("Ingrese deserializar y comenzar juego\n")
	fmt.Println("----------------------------------------------------------------")
	fmt.Println("----------------------------------------------------------------")
	gameOver = make(chan struct{})
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
				if player.Position == 15 {
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

	fmt.Printf("Jugador %d (equipo %d) juega %s. Jugador %d (equipo %d) juega %s.\n",
		player1.ID, player1.Teamint, handSigns[hand1],
		player2.ID, player2.Teamint, handSigns[hand2])

	estadoActual := fmt.Sprintf("Jugador %d (equipo %d) juega %s. Jugador %d (equipo %d) juega %s.\n",
		player1.ID, player1.Teamint, handSigns[hand1],
		player2.ID, player2.Teamint, handSigns[hand2])

	enviar(hostLocales[player1.Teamint], estadoActual)
	enviar(hostLocales[player2.Teamint], estadoActual)

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
			enviar(hostLocales[player1.Teamint], fmt.Sprintf("Jugador %d (equipo %d) pierde.\n", player1.ID, player1.Teamint))
			enviar(hostLocales[player2.Teamint], fmt.Sprintf("Jugador %d (equipo %d) gana.\n", player2.ID, player2.Teamint))
		} else { // Tijeras
			player1.Position++
			player2.Position = 0
			enviar(hostLocales[player1.Teamint], fmt.Sprintf("Jugador %d (equipo %d) gana.\n", player1.ID, player1.Teamint))
			enviar(hostLocales[player2.Teamint], fmt.Sprintf("Jugador %d (equipo %d) pierde.\n", player2.ID, player2.Teamint))
		}
	case 1: // Papel
		if hand2 == 0 { // Piedra
			player1.Position++
			player2.Position = 0
			enviar(hostLocales[player1.Teamint], fmt.Sprintf("Jugador %d (equipo %d) gana.\n", player1.ID, player1.Teamint))
			enviar(hostLocales[player2.Teamint], fmt.Sprintf("Jugador %d (equipo %d) pierde.\n", player2.ID, player2.Teamint))
		} else { // Tijeras
			player2.Position++
			player1.Position = 0
			enviar(hostLocales[player1.Teamint], fmt.Sprintf("Jugador %d (equipo %d) pierde.\n", player1.ID, player1.Teamint))
			enviar(hostLocales[player2.Teamint], fmt.Sprintf("Jugador %d (equipo %d) gana.\n", player2.ID, player2.Teamint))
		}
	case 2: // Tijeras
		if hand2 == 0 { // Piedra
			player2.Position++
			player1.Position = 0
			enviar(hostLocales[player1.Teamint], fmt.Sprintf("Jugador %d (equipo %d) pierde.\n", player1.ID, player1.Teamint))
			enviar(hostLocales[player2.Teamint], fmt.Sprintf("Jugador %d (equipo %d) gana.\n", player2.ID, player2.Teamint))
		} else { // Papel
			player1.Position++
			player2.Position = 0
			enviar(hostLocales[player1.Teamint], fmt.Sprintf("Jugador %d (equipo %d) gana.\n", player1.ID, player1.Teamint))
			enviar(hostLocales[player2.Teamint], fmt.Sprintf("Jugador %d (equipo %d) pierde.\n", player2.ID, player2.Teamint))
		}
	}

	player1.Mutex.Unlock()
	player2.Mutex.Unlock()
}

func getWinningTeam(teams []Team) Team {
	maxPoints := 0
	var winningTeam Team

	for _, team := range teams {
		if team.Points > maxPoints {
			maxPoints = team.Points
			winningTeam = team
		}
	}

	return winningTeam
}

func myIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Fatal(err)
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}

	return ""
}
