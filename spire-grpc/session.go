package spire_grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffegrpc/grpccredentials"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
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

func NewSpireClient(spireServer string, trustDomain string, uds string) (*SPIREClient, error) {
	// Create a new SPIRE client using the SPIFFE Workload API
	logger := logrus.New()
	logger.Info("Creating new spire source...")
	ctx := context.Background()
	source, err := workloadapi.NewX509Source(ctx,
		workloadapi.WithClientOptions(workloadapi.WithAddr(fmt.Sprintf("unix://%s", uds))))
	if err != nil {
		logger.Errorf("Failed to create X509 source: %v", err)
		return nil, err
	}

	logger.Info("Creating new connection...")
	// MTLS connection to SPIRE server
	serverID := spiffeid.RequireFromString(fmt.Sprintf("spiffe://%s/spire/server", trustDomain))
	conn, err := grpc.NewClient(spireServer, grpc.WithTransportCredentials(
		grpccredentials.MTLSClientCredentials(source, source, tlsconfig.AuthorizeID(serverID))))
	if err != nil {
		logger.Errorf("Failed to create gRPC connection: %v", err)
		return nil, err
	}

	sc := &SPIREClient{
		Logger:   logrus.New(),
		GRPCConn: conn,
		Client:   entrypb.NewEntryClient(conn),
	}

	return sc, nil
}

func NewClient(spireServer string) (*SPIREClient, error) {
	// Create a new SPIRE client using cert and key files
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

	conn, err := grpc.NewClient(spireServer, grpc.WithTransportCredentials(grpcCreds))

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
