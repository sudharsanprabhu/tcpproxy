package main

import (
	"log"
	"net"
	"tcpproxy/internal/tcp"
)

const (
	LOCAL_SERVER = "localhost:4002"
	REMOTE_SERVER = "192.168.1.2:5000"
	CONTROL_SERVER = "192.168.1.2:8000"
)

const TOKEN = "token2"

func main() {
	control := tcp.Connect(CONTROL_SERVER)

	if _, err := control.Write([]byte(TOKEN + "\n")); err != nil {
		log.Fatalf("[%s] %v", CONTROL_SERVER, err)
	}
	
	for {
		listenNotification(control)
		go proxy(LOCAL_SERVER, REMOTE_SERVER)
	}
}

func listenNotification(conn net.Conn) {
	buffer := make([]byte, 1)
	_, err := conn.Read(buffer)
	if err != nil {
		log.Printf("[%s] %v", conn.RemoteAddr(), err)
		return
	}
}

func proxy(localAddress string, remoteAddress string) {
	localConn := tcp.Connect(localAddress)
	remoteConn := tcp.Connect(remoteAddress)
	tcp.Pipe(localConn, remoteConn)
}