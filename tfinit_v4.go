package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"strconv"
	"bufio"
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

var players []Player
var teams []Team
var numTeam int

var htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
	<title>Team Configuration</title>
</head>
<body>
	<h1>Team Configuration</h1>
	<form action="/configure" method="POST">
		<label for="numTeams">Number of Teams:</label>
		<input type="number" id="numTeams" name="numTeams" min="1" required><br>

		<label for="numPlayersPerTeam">Number of Players per Team:</label>
		<input type="number" id="numPlayersPerTeam" name="numPlayersPerTeam" min="1" required><br>

		<input type="submit" value="Submit">
	</form>
</body>
</html>
`

func main() {
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/configure", handleConfigure)
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal(err)
		}
	}()

	//remote_Host := fmt.Sprintf("%s:%d", "192.168.9.48", 8002)
	//conn, _ := net.Dial("tcp", remote_Host)
	//defer conn.Close()

	//remote_Host_Notifica := fmt.Sprintf("%s:%d", "192.168.9.48", 8001)
	//conn_Notifica, _ := net.Dial("tcp", remote_Host_Notifica)
	//defer conn_Notifica.Close()

	select {} // Keep the program running
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("configure").Parse(htmlTemplate))
	if err := tmpl.Execute(w, nil); err != nil {
		log.Println(err)
	}
}

func handleConfigure(w http.ResponseWriter, r *http.Request) {
	numTeamsStr := r.FormValue("numTeams")
	numPlayersPerTeamStr := r.FormValue("numPlayersPerTeam")

	numTeams, err := strconv.Atoi(numTeamsStr)
	if err != nil {
		http.Error(w, "Invalid number of teams", http.StatusBadRequest)
		return
	}

	numPlayersPerTeam, err := strconv.Atoi(numPlayersPerTeamStr)
	if err != nil {
		http.Error(w, "Invalid number of players per team", http.StatusBadRequest)
		return
	}

	teams := make([]Team, numTeams)
	for i := 0; i < numTeams; i++ {
		teams[i] = Team{ID: i, PointsTarget: numPlayersPerTeam, IDmint: numTeam}
	}

	players := make([]Player, numTeams*numPlayersPerTeam)
	for i := 0; i < numTeams; i++ {
		for j := 0; j < numPlayersPerTeam; j++ {
			playerID := i*numPlayersPerTeam + j
			players[playerID] = Player{
				ID:       playerID % numPlayersPerTeam,
				Team:     i,
				Position: 0,
				Meta:     false,
				Teamint:  i,
			}
		}
	}

	arryJsonPlayer, _ := json.Marshal(players)
	jsonStringPlayer := string(arryJsonPlayer)

	mensaje := Mensaje{NumTeams: numTeams, IndexSearch: 0}
	jsonMensaje, _ := json.Marshal(mensaje)

	fmt.Fprintf(w, "Configuration Successful<br>")
	fmt.Fprintf(w, "Number of Teams: %d<br>", numTeams)
	fmt.Fprintf(w, "Number of Players per Team: %d<br>", numPlayersPerTeam)
	fmt.Fprintf(w, "<br>")

	fmt.Fprintf(w, "Teams:<br>")
	for _, team := range teams {
		fmt.Fprintf(w, "Team ID: %d, Points Target: %d<br>", team.ID, team.PointsTarget)
	}
	fmt.Fprintf(w, "<br>")

	fmt.Fprintf(w, "Players:<br>")
	for _, player := range players {
		fmt.Fprintf(w, "Player ID: %d, Team: %d, Position: %d<br>", player.ID, player.Team, player.Position)
	}

	remote_Host := fmt.Sprintf("%s:%d", "192.168.9.48", 8002)
	conn, _ := net.Dial("tcp", remote_Host)
	defer conn.Close()

	remote_Host_Notifica := fmt.Sprintf("%s:%d", "192.168.9.48", 8001)
	conn_Notifica, _ := net.Dial("tcp", remote_Host_Notifica)
	defer conn_Notifica.Close()

	fmt.Fprintf(conn_Notifica, "%s\n%s\n", "Informacion", string(jsonMensaje))
	fmt.Fprintf(conn, "%s\n%s\n", "All", jsonStringPlayer)
}
