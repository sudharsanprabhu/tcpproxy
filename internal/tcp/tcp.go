package tcp

import (
	"io"
	"net"
	"sync"
)


func Pipe(conn1 net.Conn, conn2 net.Conn) {
	defer conn1.Close()
	defer conn2.Close()

	var waitGroup sync.WaitGroup
	waitGroup.Add(2)

	go func() {
		io.Copy(conn1, conn2)
		waitGroup.Done()
	}()

	go func() {
		io.Copy(conn2, conn1)
		waitGroup.Done()
	}()

	waitGroup.Wait()
}
