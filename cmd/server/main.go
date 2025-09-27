package main

import (
	"bufio"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"tcpproxy/internal/tcp"
	"tcpproxy/cmd/server/internal/config"
)


type ControlClientRegistry struct {
	clients map[string]net.Conn
	mutex sync.RWMutex
}

var controlRegistry = ControlClientRegistry {
	clients:  make(map[string]net.Conn),
	mutex: sync.RWMutex{},
}


func main() {
	config, err := config.LoadConfig("config.toml")
	if err != nil {
		log.Fatal(err)
	}

	clientListener := tcp.Listen(config.ClientPort)
	defer clientListener.Close()

	controlListener := tcp.Listen(config.ControlPort)
	go acceptControls(controlListener)

	for _, server := range config.Servers {
		go proxy(server, clientListener)
	}

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	<-signalChannel
}

func acceptControls(listener net.Listener) {
	defer listener.Close()
	for {
		socket, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go func(socket net.Conn) {
			reader := bufio.NewReader(socket)
			token, err := reader.ReadString('\n')
			if err != nil {
				log.Printf("[CONTROL] [%s] %v", socket.RemoteAddr(), err)
				return
			}

			token = strings.TrimSpace(token)
			controlRegistry.mutex.Lock()
			controlRegistry.clients[token] = socket
			controlRegistry.mutex.Unlock()
		}(socket)
	}
}

func proxy(server config.ProxyServer, clientListener net.Listener) {
	listener := tcp.Listen(server.Port)
	defer listener.Close()

	for {
		publicSocket, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		
		go func(publicSocket net.Conn) {
			clientSocket, err := getClient(server.Token, clientListener)
			if err != nil {
				log.Printf("[%s] Error getting client for %s: %v", server.Name, publicSocket.RemoteAddr(), err)
				publicSocket.Close()
				return
			}

			tcp.Pipe(publicSocket, clientSocket)
		}(publicSocket)
	}
}

func getClient(token string, clientListener net.Listener) (net.Conn, error) {
	controlRegistry.mutex.RLock()
	control, ok := controlRegistry.clients[token]
	controlRegistry.mutex.RUnlock()
	if !ok {
		return nil, errors.New("Control not found for " + token)
	}

	_, err := control.Write([]byte{1})
	if err != nil {
		return nil, err
	}

	clientSocket, err := clientListener.Accept()
	if err != nil {
		return nil, err
	}

	return clientSocket, nil
}
