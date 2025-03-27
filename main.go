package main

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/sirupsen/logrus"
	entrypb "github.com/spiffe/spire-api-sdk/proto/spire/api/server/entry/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"os"
	"spire-api/spire-api"
)

const (
	CERT = "spire-api.crt"
	KEY  = "spire-api.key"
	CA   = "ca.crt"
)

func main() {
	logger := logrus.New()
	logger.Info("Hello, World!")
	spireServer := "omegaspire01.omegaworld.net:8081"

	caCert, err := os.ReadFile(CA)
	if err != nil {
		logger.Errorf("Failed to read CA file: %v", err)
		return
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		logger.Errorf("Failed to append CA file")
		return
	}
	clientCert, err := tls.LoadX509KeyPair(CERT, KEY)
	if err != nil {
		logger.Errorf("Failed to load client cert: %v", err)
		return
	}
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{clientCert},
		RootCAs:            certPool,
		InsecureSkipVerify: true,
	}
	grpcCreds := credentials.NewTLS(tlsConfig)

	logger.Infof("Connecting to SPIRE server: %v", spireServer)

	conn, err := grpc.Dial(spireServer, grpc.WithTransportCredentials(grpcCreds))

	if err != nil {
		logger.Errorf("Failed to connect to SPIRE server: %v", err)
		return
	}
	defer conn.Close()
	logger.Info("Connected to SPIRE server")

	sc := spire_api.SPIREClient{
		Logger:   logrus.New(),
		GRPCConn: conn,
		Client:   entrypb.NewEntryClient(conn),
	}

	e := &spire_api.Entry{
		TrustDomain:    "wl.dev.omegaworld.net",
		ServiceAccount: "dummy3",
		Namespace:      "myapp",
		Cluster:        "ambient-b",
	}
	sc.GetEntryByID("b4bd7e0d-cb1d-4a93-bb8b-fe8b5314e0ae")
	sc.CreateEntry(e)

}
