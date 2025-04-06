package main

import (
	"flag"
	"github.com/sirupsen/logrus"
	server "spire-api/api"
)

func main() {
	//add argument for server address
	serverAddress := flag.String("server", "omegaspire01.omegaworld.net", "SPIRE server address")
	serverPort := flag.Int("port", 8081, "SPIRE server port")
	apiPort := flag.Int("api-port", 8080, "API server port")
	spireDir := flag.String("spire-dir", "/opt/spire", "SPIRE directory path")
	flag.Parse()

	logger := logrus.New()
	logger.Info("Calling Start...")
	server.Start(*serverAddress, *serverPort, *apiPort, *spireDir)
}
