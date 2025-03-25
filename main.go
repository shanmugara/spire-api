package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	entrypb "github.com/spiffe/spire-api-sdk/proto/spire/api/server/entry/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"spire-api/spire-api"
)

func main() {
	logger := logrus.New()
	logger.Info("Hello, World!")
	fmt.Println("Hello, World!")
	spireServer := "omegaspire01.omegaworld.net:8081"
	logger.Infof("Connecting to SPIRE server: %v", spireServer)
	conn, err := grpc.Dial(spireServer, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Errorf("Failed to connect to SPIRE server: %v", err)
		return
	}
	defer conn.Close()
	logger.Info("Connected to SPIRE server")

	sc := spire_api.SPIREClient{
		Logger:   logrus.New(),
		GPRCConn: conn,
		Client:   entrypb.NewEntryClient(conn),
	}
	sc.GetEntries()

}
