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
	KubeConfig     string `json:"kubeConfig,omitempty"` // Optional, used for KubeConfig entries
}

type SPIREClient struct {
	Logger   *logrus.Logger
	GRPCConn *grpc.ClientConn
	Client   entrypb.EntryClient
}

// create structs for HCL mapping

type Plugins struct {
	Notifier     Notifier     `hcl:"notifier"`
	NodeAttestor NodeAttestor `hcl:"node_attestor"`
}

type Notifier struct {
	Type       string           `hcl:"type,label"` // e.g., "k8sbundle"
	PluginData NotifyPluginData `hcl:"plugin_data,block"`
}

type NotifyPluginData struct {
	PluginData NotifyClusters `hcl:"plugin_data,block"`
}
type NotifyClusters struct {
	Clusters []NotifyClusterConfig `hcl:"cluster"`
}

type NotifyClusterConfig struct {
	Namespace      string `hcl:"namespace"`
	KubeconfigPath string `hcl:"kube_config_file_path"`
}

type NodeAttestor struct {
	Name       string         `hcl:"name,label"`
	PluginData PSATPluginData `hcl:"plugin_data, block"`
}

type PSATPluginData struct {
	Clusters map[string]PSATClusterConfig `hcl:"clusters"`
}

type PSATClusterConfig struct {
	ServiceAccountAllowList []string `hcl:"service_account_allow_list"`
	KubeConfigFile          string   `hcl:"kube_config_file"`
}
