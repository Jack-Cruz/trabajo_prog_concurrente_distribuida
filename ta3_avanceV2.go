package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Player struct {
	ID       int
	Team     int
	Position int
	Points   int
	Mutex    sync.Mutex
}

type Team struct {
	ID           int
	Points       int
	PointsTarget int
	Mutex        sync.Mutex
}

func main() {
	rand.Seed(time.Now().UnixNano())

	numTeams := 3
	numPlayersPerTeam := 4
	pointsTarget := 20

	teams := make([]Team, numTeams)
	for i := 0; i < numTeams; i++ {
		teams[i] = Team{ID: i, PointsTarget: int(float64(pointsTarget) * 1.5)}
	}

	players := make([]Player, numTeams*numPlayersPerTeam)
	for i := 0; i < numTeams; i++ {
		for j := 0; j < numPlayersPerTeam; j++ {
			playerID := i*numPlayersPerTeam + j
			players[playerID] = Player{
				ID:       playerID,
				Team:     i,
				Position: 0,
				Points:   0,
			}
		}
	}

	gameOver := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(len(players))

	for i := 0; i < len(players); i++ {
		go func(player *Player) {
			defer wg.Done()
			for {
				player.Mutex.Lock()
				player.Position++
				player.Mutex.Unlock()

				time.Sleep(time.Millisecond * time.Duration(rand.Intn(500)))

				opponent := getOpponent(players, player)
				if opponent != nil && opponent.Position == player.Position {
					playRockPaperScissors(player, opponent)
				}

				player.Mutex.Lock()
				teams[player.Team].Mutex.Lock()
				teams[player.Team].Points += player.Points
				teams[player.Team].Mutex.Unlock()
				player.Mutex.Unlock()

				if teams[player.Team].Points >= teams[player.Team].PointsTarget {
					close(gameOver)
					return
				}
			}
		}(&players[i])
	}

	startTime := time.Now() // Obtener el tiempo de inicio

	go func() {
		wg.Wait()
		close(gameOver)
	}()

	select {
	case <-gameOver:
		winningTeam := getWinningTeam(teams)
		fmt.Printf("¡El equipo %d ha ganado con %d puntos!\n", winningTeam.ID, winningTeam.Points)

		endTime := time.Now()                     // Obtener el tiempo actual
		duration := endTime.Sub(startTime)        // Calcular la duración
		fmt.Printf("Duración de la ejecución: %s\n", duration)
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

	fmt.Printf("Jugador %d (equipo %d) juega %s puntos %d. Jugador %d (equipo %d) juega %s puntos %d.\n",
		player1.ID, player1.Team, handSigns[hand1], player1.Points,
		player2.ID, player2.Team, handSigns[hand2], player2.Points)

	if hand1 == hand2 {
		return
	}

	player1.Mutex.Lock()
	player2.Mutex.Lock()

	switch hand1 {
	case 0: // Piedra
		if hand2 == 1 { // Papel
			player2.Position++
			player2.Points++
			player1.Position--

			if player1.Points > 0 {
				player1.Points--
			}
		} else { // Tijeras
			player1.Position++
			player1.Points++
			player2.Position--
			//player2.Points--
			if player2.Points > 0 {
				player2.Points--
			}
		}
	case 1: // Papel
		if hand2 == 0 { // Piedra
			//player1.Position--
			player1.Points--
			player2.Position++
			player2.Points++
			if player1.Points > 0 {
				player1.Points--
			}
		} else { // Tijeras
			player2.Position--
			//player2.Points--
			player1.Position++
			player1.Points++
			if player2.Points > 0 {
				player2.Points--
			}
		}
	case 2: // Tijeras
		if hand2 == 0 { // Piedra
			player2.Position++
			player2.Points++
			player1.Position--
			//player1.Points--
			if player1.Points > 0 {
				player1.Points--
			}
		} else { // Papel
			player1.Position++
			player1.Points++
			player2.Position--
			//player2.Points--
			if player2.Points > 0 {
				player2.Points--
			}
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