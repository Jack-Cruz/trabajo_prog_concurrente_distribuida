package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

var addressLocal string

func main() {
	//Lectura por consola del host origin
	brInput := bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese el puerto del host local: ")
	puertoLocal, _ := brInput.ReadString('\n')
	puertoLocal = strings.TrimSpace(puertoLocal)

	addressLocal = fmt.Sprintf("localhost:%s", puertoLocal)

	enviar(0)
}

func enviar(n int) {
	conn, _ := net.Dial("tcp", addressLocal)

	str := ""
	defer conn.Close()
	fmt.Fprintf(conn, "%d\n%s\n%s\n", n, str, str)

}
