package spire_grpc

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/sirupsen/logrus"
	entrypb "github.com/spiffe/spire-api-sdk/proto/spire/api/server/entry/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"os"
)

const (
	CERT = "certs/spire-api.crt"
	KEY  = "certs/spire-api.key"
	CA   = "certs/ca.crt"
)

func NewClient(spireServer string) (*SPIREClient, error) {
	logger := logrus.New()

	caCert, err := os.ReadFile(CA)
	if err != nil {
		logger.Errorf("Failed to read CA file: %v", err)
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		logger.Errorf("Failed to append CA file")
		return nil, err
	}
	clientCert, err := tls.LoadX509KeyPair(CERT, KEY)
	if err != nil {
		logger.Errorf("Failed to load client cert: %v", err)
		return nil, err
	}
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{clientCert},
		RootCAs:            certPool,
		InsecureSkipVerify: true,
	}
	grpcCreds := credentials.NewTLS(tlsConfig)

	logger.Infof("Creating connection to SPIRE server: %v", spireServer)

	conn, err := grpc.Dial(spireServer, grpc.WithTransportCredentials(grpcCreds))

	if err != nil {
		logger.Errorf("Failed to create connection to SPIRE server: %v", err)
		return nil, err
	}

	logger.Info("Connection created to SPIRE server")

	sc := &SPIREClient{
		Logger:   logrus.New(),
		GRPCConn: conn,
		Client:   entrypb.NewEntryClient(conn),
	}
	return sc, nil
}
