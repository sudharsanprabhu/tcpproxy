package main

import (
	"net"
	"os"
	"tcpproxy/internal/logger"
	"tcpproxy/internal/tcp"

	"github.com/joho/godotenv"
)


var log = logger.GetLogger("client")

func main() {
	godotenv.Load();
	LOCAL_SERVER := os.Getenv("LOCAL_SERVER")
	REMOTE_SERVER := os.Getenv("REMOTE_SERVER")
	CONTROL_SERVER := os.Getenv("CONTROL_SERVER")
	TOKEN := os.Getenv("TOKEN")

	control, err := net.Dial("tcp", CONTROL_SERVER)
	if err != nil {
		log.Fatalf("[%s] Connection failed - %v", CONTROL_SERVER, err)
	}
	log.Infof("[%s] Connected", CONTROL_SERVER)

	if _, err := control.Write([]byte(TOKEN + "\n")); err != nil {
		log.Fatalf("[%s] %v", CONTROL_SERVER, err)
	}
	
	for {
		isNotified := listenNotification(control)
		if isNotified {
			go proxy(LOCAL_SERVER, REMOTE_SERVER)
		}
	}
}

func listenNotification(conn net.Conn) bool {
	buffer := make([]byte, 1)
	_, err := conn.Read(buffer)
	if err != nil {
		log.Fatalf("[%s] %v", conn.RemoteAddr(), err)
		return false
	}

	return true
}

func proxy(localAddress string, remoteAddress string) {
	localConn, err := net.Dial("tcp", localAddress)
	if err != nil {
		log.Fatalf("[%s] Connection failed - %v", localAddress, err)
	}
	log.Infof("[%s] Connected", localAddress)

	remoteConn, err := net.Dial("tcp", remoteAddress)
	if err != nil {
		log.Fatalf("[%s] Connection failed - %v", remoteAddress, err)
	}
	log.Infof("[%s] Connected", remoteAddress)

	log.Infof("Piping [%s] <-> [%s]", localAddress, remoteAddress)
	tcp.Pipe(localConn, remoteConn)
	log.Infof("Pipe closed [%s] <-> [%s]", localAddress, remoteAddress)
}