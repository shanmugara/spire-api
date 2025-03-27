package spire_grpc

import (
	"context"
	"fmt"
	entrypb "github.com/spiffe/spire-api-sdk/proto/spire/api/server/entry/v1"
	"github.com/spiffe/spire-api-sdk/proto/spire/api/types"
)

const (
	AgentNamespace      = "spire"
	AgentServiceAccount = "spire-agent"
	ParentRoot          = "/spire/server"
	NS                  = "ns"
	SA                  = "sa"
	KeyAgentNS          = "agent_ns"
	KeyAgentSA          = "agent_sa"
)

func (sc *SPIREClient) GetEntries() ([]*types.Entry, error) {
	resp, err := (sc.Client).ListEntries(context.Background(), &entrypb.ListEntriesRequest{})
	if err != nil {
		sc.Logger.Errorf("Failed to list entries: %v", err)
		return nil, fmt.Errorf(err.Error())
	}
	for _, entry := range resp.Entries {
		sc.Logger.Infof("Entry: %v", entry)
	}
	return resp.Entries, nil
}

func (sc *SPIREClient) GetEntryByID(id string) {
	resp, err := (sc.Client).GetEntry(context.Background(), &entrypb.GetEntryRequest{Id: id})
	if err != nil {
		sc.Logger.Errorf("Failed to get entry: %v", err)
		return
	}
	sc.Logger.Infof("Entry: %v", resp.SpiffeId)
}

func (sc *SPIREClient) GetEntryBySPIFFE(e *Entry) {
	sc.Logger.Infof("fetching entry by spiffeID")
	spiffeID := &types.SPIFFEID{
		TrustDomain: e.TrustDomain,
		Path:        fmt.Sprintf("/ns/%s/sa/%s", e.Namespace, e.ServiceAccount),
	}
	req := &entrypb.ListEntriesRequest{
		Filter: &entrypb.ListEntriesRequest_Filter{
			BySpiffeId: spiffeID,
		},
	}
	resp, err := (sc.Client).ListEntries(context.Background(), req)
	if err != nil {
		sc.Logger.Errorf("Error listing entry by spiffeid %s", err.Error())
	}
	sc.Logger.Infof("%s", resp.Entries)
}

func (sc *SPIREClient) CreateEntry(e *Entry) {
	sc.Logger.Infof("Creating entry")

	pPath := fmt.Sprintf("/ns/%s/sa/%s", AgentNamespace, AgentServiceAccount)
	nsKey := NS
	saKey := SA
	if e.ServiceAccount == AgentServiceAccount && e.Namespace == AgentNamespace {
		pPath = ParentRoot
		nsKey = KeyAgentNS
		saKey = KeyAgentSA
	}

	var sel []*types.Selector

	sel = append(sel,
		&types.Selector{
			Type:  "k8s_psat",
			Value: fmt.Sprintf("cluster:%s", e.Cluster),
		},
		&types.Selector{
			Type:  "k8s_psat",
			Value: fmt.Sprintf("%s:%s", nsKey, e.Namespace),
		},
		&types.Selector{
			Type:  "k8s_psat",
			Value: fmt.Sprintf("%s:%s", saKey, e.ServiceAccount),
		},
	)

	entry := &entrypb.BatchCreateEntryRequest{
		Entries: []*types.Entry{
			{
				ParentId: &types.SPIFFEID{
					TrustDomain: e.TrustDomain,
					Path:        pPath,
				},
				SpiffeId: &types.SPIFFEID{
					TrustDomain: e.TrustDomain,
					Path:        fmt.Sprintf("/ns/%s/sa/%s", e.Namespace, e.ServiceAccount),
				},
				Selectors: sel,
			},
		},
	}

	resp, err := (sc.Client).BatchCreateEntry(context.Background(), entry)
	if err != nil {
		sc.Logger.Errorf("Failed to create entry: %v", err)
		return
	}
	sc.Logger.Infof("Entry: %v", resp.Results)
}
