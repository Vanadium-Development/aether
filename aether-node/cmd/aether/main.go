package main

import (
	"node/internal/node"

	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		PadLevelText:    true,
		TimestampFormat: "2006-01-02 : 15:04:05",
	})
	logrus.SetLevel(logrus.DebugLevel)
	node.InitializeNode()
}
