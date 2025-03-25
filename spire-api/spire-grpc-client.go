package spire_api

import (
	"context"
	"github.com/sirupsen/logrus"
	entrypb "github.com/spiffe/spire-api-sdk/proto/spire/api/server/entry/v1"
	"google.golang.org/grpc"
)

type SPIREClient struct {
	Logger   *logrus.Logger
	GPRCConn *grpc.ClientConn
	Client   entrypb.EntryClient
}

func (sc *SPIREClient) GetEntries() {
	resp, err := (sc.Client).ListEntries(context.Background(), &entrypb.ListEntriesRequest{})
	if err != nil {
		sc.Logger.Errorf("Failed to list entries: %v", err)
		return
	}
	for _, entry := range resp.Entries {
		sc.Logger.Infof("Entry: %v", entry)
	}
}
