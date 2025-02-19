package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
	"strings"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type todoList struct {
	//id          int
	name        string
	connections []*websocket.Conn
}

type dataT map[int]todoList

var data = make(dataT)

func getCommand(input string) string {
	inputArray := strings.Split(input, ":")
	return inputArray[0]
}

func getMessage(input string) string {
	inputArray := strings.Split(input, ":")
	return strings.Join(inputArray[1:], ":")
}

func connectionExists(connections []*websocket.Conn, connection *websocket.Conn) bool {
	for _, conn := range connections {
		if conn == connection {
			return true
		}
	}
	return false
}

func addConnection(connections []*websocket.Conn, connection *websocket.Conn) []*websocket.Conn {
	if !connectionExists(connections, connection) {
		connections = append(connections, connection)
	}
	return connections
}

func closeConnection(dataPointer *dataT, id int, connection *websocket.Conn) {
	newConnections := []*websocket.Conn{}
	for _, conn := range (*dataPointer)[id].connections {
		if conn != connection {
			newConnections = append(newConnections, conn)
		}
	}
	dataPointer[id].connections = newConnections
	connection.Close()
}

func main() {
	http.HandleFunc("/ws", func(writer http.ResponseWriter, request *http.Request) {
		id, err := strconv.Atoi(request.URL.Query().Get("id"))
		if err != nil {
			fmt.Println("ID reading error:", err)
		}
		fmt.Println("ID:", id)
		// Upgrade upgrades the HTTP server connection to the WebSocket protocol.
		connection, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			fmt.Println("Upgrade failed: ", err)
			return
		}
		//defer connection.Close()
		defer closeConnection(&data, id, connection)

		// Continuously read and write message
		for {
			_, message, err := connection.ReadMessage()
			//Nothing(messageType)
			if err != nil {
				fmt.Println("read failed: ", err)
			}
			input := string(message)
			receiveCommand := getCommand(input)
			receiveMessage := getMessage(input)
			sendCommand := ""
			sendMessage := ""
			if receiveCommand == "init" {
				fmt.Println("init")
				data[id] = todoList{
					name:        data[id].name,
					connections: addConnection(data[id].connections, connection),
				}

				sendCommand = "init"
				sendMessage = data[id].name
			}
			if receiveCommand == "name-changed" {
				fmt.Println("name has changed", receiveMessage)
				data[id] = todoList{
					//id:          id,
					name:        receiveMessage,
					connections: addConnection(data[id].connections, connection),
				}

				sendCommand = "name-changed"
				sendMessage = data[id].name
			}
			fmt.Println("CLIENTS", data[id].connections)
			for i, conn := range data[id].connections {
				//err := conn.WriteMessage(messageType, message)
				err := conn.WriteMessage(1, []byte(sendCommand+":"+sendMessage))
				if err != nil {
					fmt.Println("write failed: ", err, i)
					break
				}
			}
		}
	})
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("server error: ", err)
	}
}
