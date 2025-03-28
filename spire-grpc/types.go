package spire_grpc

import (
	"github.com/sirupsen/logrus"
	entrypb "github.com/spiffe/spire-api-sdk/proto/spire/api/server/entry/v1"
	"google.golang.org/grpc"
)

type Entry struct {
	TrustDomain    string `json:"trustDomain" required:"true"`
	ServiceAccount string `json:"serviceAccount" required:"true"`
	Namespace      string `json:"namespace" required:"true"`
	Cluster        string `json:"cluster" required:"true"`
}

type SPIREClient struct {
	Logger   *logrus.Logger
	GRPCConn *grpc.ClientConn
	Client   entrypb.EntryClient
}
