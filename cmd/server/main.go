package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"tcpproxy/cmd/server/internal/config"
	"tcpproxy/internal/logger"
	"tcpproxy/internal/tcp"
)


type ControlClientRegistry struct {
	clients map[string]net.Conn
	mutex sync.RWMutex
}

var controlRegistry = ControlClientRegistry {
	clients:  make(map[string]net.Conn),
	mutex: sync.RWMutex{},
}

var log = logger.GetLogger("server")


func main() {
	config, err := config.LoadConfig("config.toml")
	if err != nil {
		log.Fatal(err)
	}

	clientListener, err := net.Listen("tcp", fmt.Sprint(":", config.ClientPort))
	if err != nil {
		log.Fatalf("Cannot listen on %d - %v", config.ClientPort, err)
	}
	log.Info("Listening on port ", config.ClientPort)
	defer clientListener.Close()

	controlListener, err := net.Listen("tcp", fmt.Sprint(":", config.ControlPort))
	if err != nil {
		log.Fatalf("Cannot listen on %d - %v", config.ClientPort, err)
	}
	log.Info("Listening on port ", config.ControlPort)
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
			log.Error(err)
			continue
		}

		go func(socket net.Conn) {
			reader := bufio.NewReader(socket)
			token, err := reader.ReadString('\n')
			if err != nil {
				log.Errorf("[CONTROL] [%s] %v", socket.RemoteAddr(), err)
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
	listener, err := net.Listen("tcp", fmt.Sprint(":", server.Port))
	if err != nil {
		log.Fatalf("Cannot listen on %d - %v", server.Port, err)
	}
	log.Info("Listening on port ", server.Port)
	defer listener.Close()

	for {
		publicSocket, err := listener.Accept()
		if err != nil {
			log.Error(err)
			continue
		}
		
		go func(publicSocket net.Conn) {
			clientSocket, err := getClient(server.Token, clientListener)
			if err != nil {
				log.Errorf("[%s] Error getting client for %s: %v", server.Name, publicSocket.RemoteAddr(), err)
				publicSocket.Close()
				return
			}

			log.Infof("Piping [%s] <-> [%s]", publicSocket.RemoteAddr(), clientSocket.RemoteAddr())
			tcp.Pipe(publicSocket, clientSocket)
			log.Infof("Pipe closed [%s] <-> [%s]", publicSocket.RemoteAddr(), clientSocket.RemoteAddr())
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
