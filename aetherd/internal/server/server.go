package server

import (
	"encoding/json"
	"net"
	"os"

	"github.com/sirupsen/logrus"
)

const socketPath = "/tmp/aetherd.sock"

func handleConnection(conn net.Conn) {
	json.NewDecoder(conn)
}

func StartDaemonServer() {
	_ = os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		logrus.Fatalf("Could not open unix socket on %s: %s\n", socketPath, err)
		return
	}
	defer listener.Close()

	logrus.Infof("Aether daemon listening on unix socket %s\n", socketPath)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logrus.Errorf("Could not accept incoming connection: %s\n", err)
			continue
		}

		go handleConnection(conn)
	}
}
