package main

import (
	"fmt"
	"net"
	"os"
	"tcpproxy/internal/logger"
	"tcpproxy/internal/tcp"
	"time"

	"github.com/joho/godotenv"
)


var log = logger.GetLogger("client")

func main() {
	godotenv.Load();

	for {
		err := run()
		if err != nil {
			log.Error(err)
		}

		log.Info("Waiting for 5 seconds")
		time.Sleep(5 * time.Second)
	}
}

func run() error {
	LOCAL_SERVER := os.Getenv("LOCAL_SERVER")
	REMOTE_SERVER := os.Getenv("REMOTE_SERVER")
	CONTROL_SERVER := os.Getenv("CONTROL_SERVER")
	TOKEN := os.Getenv("TOKEN")

	control, err := net.Dial("tcp", CONTROL_SERVER)
	if err != nil {
		return fmt.Errorf("[%s] Connection failed - %v", CONTROL_SERVER, err)
	}
	log.Infof("[%s] Connected", CONTROL_SERVER)
	defer control.Close()

	if _, err := control.Write([]byte(TOKEN + "\n")); err != nil {
		return fmt.Errorf("[%s] %v", CONTROL_SERVER, err)
	}
	
	for {
		err := listenNotification(control)
		if err != nil {
			return err
		}
		
		go proxy(LOCAL_SERVER, REMOTE_SERVER)
	}
}

func listenNotification(conn net.Conn) error {
	buffer := make([]byte, 1)
	_, err := conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("[%s] %v", conn.RemoteAddr(), err)
	}

	return nil
}

func proxy(localAddress string, remoteAddress string) {
	localConn, err := net.Dial("tcp", localAddress)
	if err != nil {
		log.Errorf("[%s] Connection failed - %v", localAddress, err)
		return
	}
	log.Infof("[%s] Connected", localAddress)
	defer localConn.Close()

	remoteConn, err := net.Dial("tcp", remoteAddress)
	if err != nil {
		log.Errorf("[%s] Connection failed - %v", remoteAddress, err)
		return
	}
	log.Infof("[%s] Connected", remoteAddress)
	defer remoteConn.Close()

	log.Infof("Piping [%s] <-> [%s]", localAddress, remoteAddress)
	tcp.Pipe(localConn, remoteConn)
	log.Infof("Pipe closed [%s] <-> [%s]", localAddress, remoteAddress)
}