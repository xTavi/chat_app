package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

var (
	pongWait = 10 * time.Second
	// pingInterval has to be less than pongWait, We cant multiply by 0.9 to get 90% of time
	// Because that can make decimals, so instead *9 / 10 to get 90%
	// The reason why it has to be less than PingRequency is becuase otherwise it will send a new Ping before getting response
	pingInterval = (pongWait * 9) / 10
)

type ClientList map[*Client]bool

type Client struct {
	connection *websocket.Conn

	manager *Manager

	egress chan Event
}

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection: conn,
		manager:    manager,
		egress:     make(chan Event),
	}
}

func (c *Client) readMessages() {
	defer func() {

		c.manager.removeClient(c)
	}()

	for {

		_, payload, err := c.connection.ReadMessage()

		if err != nil {

			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			}
			break
		}

		var request Event
		if err := json.Unmarshal(payload, &request); err != nil {
			log.Printf("error marshalling message: %v", err)
			break
		}

		if err := c.manager.routeEvent(request, c); err != nil {
			log.Println("Error handeling Message: ", err)
		}
	}
}

func (c *Client) writeMessages() {
	defer func() {

		c.manager.removeClient(c)
	}()

	for {
		select {
		case message, ok := <-c.egress:

			if !ok {

				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {

					log.Println("connection closed: ", err)
				}

				return
			}
			data, err := json.Marshal(message)
			if err != nil {
				log.Println(err)
				return
			}

			if err := c.connection.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Println(err)
			}
			log.Println("sent message")
		}

	}
}
