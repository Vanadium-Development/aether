package server

import (
	"aetherd/internal/constants"
	"aetherd/pkg/dtos"
	"encoding/json"
	"net"
	"os"

	"github.com/sirupsen/logrus"
)

func handleConnection(conn net.Conn) {
	var dto dtos.DaemonCommonDto
	err := json.NewDecoder(conn).Decode(&dto)
	if err != nil {
		logrus.Errorf("Could not decode request json: %s\n", err)
		return
	}
}

func StartDaemonServer() {
	_ = os.Remove(constants.SocketPath)

	listener, err := net.Listen("unix", constants.SocketPath)
	if err != nil {
		logrus.Fatalf("Could not open unix socket on %s: %s\n", constants.SocketPath, err)
		return
	}
	defer listener.Close()

	logrus.Infof("Aether daemon listening on unix socket %s\n", constants.SocketPath)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logrus.Errorf("Could not accept incoming connection: %s\n", err)
			continue
		}

		go handleConnection(conn)
	}
}
