package main

import (
	"log"
	"net"
	"tcpproxy/internal/tcp"
	"github.com/joho/godotenv"
	"os"
)


func main() {
	godotenv.Load();
	LOCAL_SERVER := os.Getenv("LOCAL_SERVER")
	REMOTE_SERVER := os.Getenv("REMOTE_SERVER")
	CONTROL_SERVER := os.Getenv("CONTROL_SERVER")
	TOKEN := os.Getenv("TOKEN")

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