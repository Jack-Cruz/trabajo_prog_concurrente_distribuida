# Trabajo Programacion Concurrente : Hoop Hula Hop Rock Paper Scissors

## *Contexto*: 
El juego describe un grupo de niños dentro de una area seleccionada con hula hulas, en donde se crean distintos grupos con N cantidades de niños, ellos pueden avanzar continuamente hasta el grupo contraria exceptuando que cuando uno de estos se encuentren con otro equipo tendran que realizar el juego de piedra papel y tijeras, en donde el ganador podra continuar y el perdedor regresara, el ganador tambien ganara 1 punto para su equipo. En este caso los el equipo que consiga Puntos Totales x 1.50, obtendra la victoria del juego.

## Diseño de Logica
**Estructura Player:**  Para poder tenerlos como un objeto o instancia a cada jugador y saber su posicion y grupo perteneciente se le han asigando estos parametros.
El jugador contiene la siguiente estructura:
- ID: identificador del jugador
- Team: Identificador del equipo al que pertenece el jugador
- Pòsition: Posición actual del jugador
- Points: Puntos obtenidos por el jugador
- Mutex: Exclusión mutua para avanzar la posición del jugador
**Estructura Team:** Se utiliza para agrupar a N cantidad de players (niños), en donde tambien estos tienen un numero de puntos consigo.
