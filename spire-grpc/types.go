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
	KubeConfig     string `json:"kubeConfig,omitempty"`
	SpireDir       string `json:"spireDir,omitempty"`
}

type SPIREClient struct {
	Logger   *logrus.Logger
	GRPCConn *grpc.ClientConn
	Client   entrypb.EntryClient
}

// create structs for SPIRE configurations for K8S and Bundle

type K8SPSATConfig struct {
	Clusters []map[string]PSATCluster `json:"clusters"`
}

type PSATCluster struct {
	ServiceAccountAllowList []string `json:"service_account_allow_list"`
	KubeConfigFile          string   `json:"kube_config_file"`
}

type K8SBundleConfig struct {
	Clusters []BundleCluster `json:"clusters"`
}

type BundleCluster struct {
	KubeConfigFilePath string `json:"kube_config_file_path"`
}
