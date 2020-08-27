package main

import (
	"log"
)

type Connections struct {
	clientsIn    map[string]chan []byte
	clientsOut   map[string]chan []byte
	addClient    chan string
	removeClient chan string
}

var hub = &Connections{
	clientsIn:    make(map[string]chan []byte),
	clientsOut:   make(map[string]chan []byte),
	addClient:    make(chan string),
	removeClient: make(chan string),
}

func (hub *Connections) Init() {
	for {
		select {
		case s := <-hub.addClient:
			log.Println("Added new client", s)
			hub.clientsIn[s] = make(chan []byte)
			hub.clientsOut[s] = make(chan []byte)
		case s := <-hub.removeClient:
			log.Println("Removed client", s)
			close(hub.clientsIn[s])
			close(hub.clientsOut[s])
			delete(hub.clientsIn, s)
			delete(hub.clientsOut, s)
		}
	}
}
