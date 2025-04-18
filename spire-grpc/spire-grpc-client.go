package spire_grpc

import (
	"context"
	"encoding/base64"
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

type entryID string

//TODO: Receive a kubeconfig in CreateEntry
//TODO: Update server.conf with the kubeconfig
//TODO: Watch for cert file updates and reload the server.conf

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

func (sc *SPIREClient) GetEntryBySPIFFE(e *Entry) ([]*types.Entry, error) {
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
		return nil, err
	}
	sc.Logger.Infof("%s", resp.Entries)
	return resp.Entries, nil
}

func (sc *SPIREClient) CreateEntry(e *Entry) (*entryID, error) {
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
		return nil, err
	}

	eID := entryID(resp.Results[0].Entry.Id)
	sc.Logger.Infof("EntryID: %v", eID)

	return &eID, nil
}

func (sc *SPIREClient) DeleteEntryBySPIFFE(e *Entry) error {
	sc.Logger.Infof("Fetching entry by spiffeID first")
	resp, err := sc.GetEntryBySPIFFE(e)
	if err != nil {
		sc.Logger.Errorf("Failed to get entry by spiffeID: %v, may be deleted. ignoring", err)
		return nil
	}
	var entryIDs []string
	for _, entry := range resp {
		entryIDs = append(entryIDs, entry.Id)
	}
	sc.Logger.Infof("Deleting entry by spiffeID")

	delresp, err := (sc.Client).BatchDeleteEntry(context.Background(), &entrypb.BatchDeleteEntryRequest{
		Ids: entryIDs,
	})
	if err != nil {
		sc.Logger.Errorf("Failed to delete entry: %v", err)
		return err
	}
	sc.Logger.Infof("Entry: %v", delresp.Results)
	return nil
}

func (sc *SPIREClient) RegisterKubeConfig(e *Entry) error {
	// Placeholder for registering kubeconfig, if needed
	if e.KubeConfig == "" {
		sc.Logger.Infof("No kubeconfig provided for entry, skipping registration")
		return nil
	}
	kcBytes := e.KubeConfig
	kcDecoded, err := base64.StdEncoding.DecodeString(kcBytes)
	if err != nil {
		sc.Logger.Errorf("Failed to decode kubeconfig: %v", err)
		return err
	}
	//DEBUG
	sc.Logger.Infof("Registering kubeconfig for entry: %v", kcDecoded)

	sc.Logger.Infof("Registering kubeconfig for entry: %v", e)
	// In a real implementation, you would call the appropriate SPIRE API to register the kubeconfig
	// For now, just return nil to indicate success
	return nil
}
