package tcp

import (
	"net"
	"log"
	"fmt"
	"sync"
	"io"
)


func Listen(port int) net.Listener {
	listener, err := net.Listen("tcp", fmt.Sprint(":", port))
	if err != nil {
		log.Fatalf("Cannot listen on %d - %v", port, err)
	}

	log.Println("Listening on port", port)
	return listener
}

func Connect(address string) net.Conn {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalf("[%s] Connection failed - %v", address, err)
	}
	
	log.Printf("[%s] Connected", address)
	return conn
}

func Pipe(conn1 net.Conn, conn2 net.Conn) {
	defer conn1.Close()
	defer conn2.Close()

	var waitGroup sync.WaitGroup
	waitGroup.Add(2)

	go func() {
		_, err := io.Copy(conn1, conn2)
		if err != nil {
			log.Println(err)
		}

		waitGroup.Done()
	}()

	go func() {
		_, err := io.Copy(conn2, conn1)
		if err != nil {
			log.Println(err)
		}

		waitGroup.Done()
	}()

	waitGroup.Wait()
}
