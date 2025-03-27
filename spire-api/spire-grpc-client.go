package spire_api

import (
	"context"
	"fmt"
	entrypb "github.com/spiffe/spire-api-sdk/proto/spire/api/server/entry/v1"
	"github.com/spiffe/spire-api-sdk/proto/spire/api/types"
)

const (
	ParentNamespace      = "spire"
	ParentServiceAccount = "spire-agent"
	ParentRoot           = "/spire/server"
	NS                   = "ns"
	SA                   = "sa"
	AgentNS              = "agent_ns"
	AgentSA              = "agent_sa"
)

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

func (sc *SPIREClient) GetEntryByID(id string) {
	resp, err := (sc.Client).GetEntry(context.Background(), &entrypb.GetEntryRequest{Id: id})
	if err != nil {
		sc.Logger.Errorf("Failed to get entry: %v", err)
		return
	}
	sc.Logger.Infof("Entry: %v", resp.SpiffeId)
}

func (sc *SPIREClient) CreateEntry(e *Entry) {
	sc.Logger.Infof("Creating entry")

	pPath := fmt.Sprintf("/ns/%s/sa/%s", ParentNamespace, ParentServiceAccount)
	ns_key := NS
	sa_key := SA
	if e.ServiceAccount == ParentServiceAccount && e.Namespace == ParentNamespace {
		pPath = ParentRoot
		ns_key = AgentNS
		sa_key = AgentSA
	}

	var sel []*types.Selector

	sel = append(sel,
		&types.Selector{
			Type:  "k8s_psat",
			Value: fmt.Sprintf("cluster:%s", e.Cluster),
		},
		&types.Selector{
			Type:  "k8s_psat",
			Value: fmt.Sprintf("%s:%s", ns_key, e.Namespace),
		},
		&types.Selector{
			Type:  "k8s_psat",
			Value: fmt.Sprintf("%s:%s", sa_key, e.ServiceAccount),
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
