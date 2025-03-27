package spire_api

import (
	"github.com/sirupsen/logrus"
	entrypb "github.com/spiffe/spire-api-sdk/proto/spire/api/server/entry/v1"
	"google.golang.org/grpc"
)

type Entry struct {
	TrustDomain    string
	ServiceAccount string
	Namespace      string
	Cluster        string
}

type SPIREClient struct {
	Logger   *logrus.Logger
	GRPCConn *grpc.ClientConn
	Client   entrypb.EntryClient
}
