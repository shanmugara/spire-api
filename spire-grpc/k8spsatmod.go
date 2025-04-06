package spire_grpc

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	k8sPsatConfigFile   = "k8s_psat.json"
	k8sBundleConfigFile = "k8s_bundle.json"
)

func (sc *SPIREClient) GetK8sPsatConfig(e *Entry) (*K8SPSATConfig, error) {
	// Read the k8s_psat config file and return the parsed K8SPSATConfig struct
	sc.Logger.Infof("Reading k8s_psat config file")
	data, err := os.ReadFile(filepath.Join(e.SpireDir, k8sPsatConfigFile))
	if err != nil {
		sc.Logger.Errorf("Failed to read k8s_psat config file: %v", err)
		return nil, err
	}
	k8spsat := &K8SPSATConfig{}
	err = json.Unmarshal(data, k8spsat)
	if err != nil {
		sc.Logger.Errorf("Failed to unmarshal k8s_psat config file: %v", err)
		return nil, err
	}
	return k8spsat, nil
}

func (sc *SPIREClient) GetK8sBundleConfig(e *Entry) (*K8SBundleConfig, error) {
	// Read the k8s_bundle config file and return the parsed K8SBundleConfig struct
	sc.Logger.Infof("Reading k8s_bundle config file")
	data, err := os.ReadFile(filepath.Join(e.SpireDir, k8sBundleConfigFile))
	if err != nil {
		sc.Logger.Errorf("Failed to read k8s_bundle config file: %v", err)
		return nil, err
	}
	k8sBundle := &K8SBundleConfig{}
	err = json.Unmarshal(data, k8sBundle)
	if err != nil {
		sc.Logger.Errorf("Failed to unmarshal k8s_bundle config file: %v", err)
		return nil, err
	}
	return k8sBundle, nil
}

func (sc *SPIREClient) MakePSATCluster(e *Entry) *PSATCluster {
	sc.Logger.Infof("Creating PSATCluster instance")
	kc := e.Cluster + ".yaml"
	pc := &PSATCluster{
		ServiceAccountAllowList: []string{"spire:spire-agent"},
		KubeConfigFile:          filepath.Join(e.SpireDir, kc), // This can be set later when creating the entry
	}
	return pc
}

func (sc *SPIREClient) MakeK8sBundleCluster(e *Entry) *BundleCluster {
	sc.Logger.Infof("Creating BundleCluster instance")
	kc := e.Cluster + ".yaml"
	bc := &BundleCluster{
		KubeConfigFilePath: filepath.Join(e.SpireDir, kc),
	}
	return bc
}

func (sc *SPIREClient) UpdateK8sPsat(e *Entry) error {
	currentPsat, err := sc.GetK8sPsatConfig(e)
	if err != nil {
		sc.Logger.Errorf("Failed to get current k8s_psat config: %v", err)
		return err
	}
	newCluster := sc.MakePSATCluster(e)
	UpdateCluster := false // Flag to check if we need to update an existing cluster

	// Check if cluster already exists in the current configuration
	for _, cluster := range currentPsat.Clusters {
		// Check if the cluster already exists in the configuration
		if _, Exists := cluster[e.Cluster]; Exists {
			sc.Logger.Infof("Cluster %s already exists in k8s_psat config, updating it", e.Cluster)
			UpdateCluster = true
			// Update the existing cluster's KubeConfigFile if needed
			cluster[e.Cluster] = *newCluster
		}
	}

	// Append the new cluster to the existing clusters
	if currentPsat != nil && UpdateCluster == false {
		sc.Logger.Infof("Appending new cluster %s to k8s_psat config", e.Cluster)
		currentPsat.Clusters = append(currentPsat.Clusters, map[string]PSATCluster{e.Cluster: *newCluster})
	}
	outFile, err := json.MarshalIndent(currentPsat, "", "  ")
	if err != nil {
		sc.Logger.Errorf("Failed to marshal updated k8s_psat config: %v", err)
		return err
	}
	// Write the updated config back to file
	if err := os.WriteFile(filepath.Join(e.SpireDir, "kubeconfigs", k8sPsatConfigFile), outFile, 0644); err != nil {
		sc.Logger.Errorf("Failed to write updated k8s_psat config file: %v", err)
		return err
	}
	sc.Logger.Infof("Successfully updated k8s_psat config file")
	return nil
}

func (sc *SPIREClient) UpdateK8sBundle(e *Entry) error {
	currentBundle, err := sc.GetK8sBundleConfig(e)
	if err != nil {
		sc.Logger.Errorf("Failed to get current k8s_bundle config: %v", err)
		return err
	}
	newCluster := sc.MakeK8sBundleCluster(e)
	// Append the new cluster to the existing clusters
	if currentBundle != nil {
		currentBundle.Clusters = append(currentBundle.Clusters, *newCluster)
	}
	outFile, err := json.MarshalIndent(currentBundle, "", "  ")
	if err != nil {
		sc.Logger.Errorf("Failed to marshal updated k8s_bundle config: %v", err)
		return err
	}
	// Write the updated config back to file
	if err := os.WriteFile(filepath.Join(e.SpireDir, "kubeconfigs", k8sBundleConfigFile), outFile, 0644); err != nil {
		sc.Logger.Errorf("Failed to write updated k8s_bundle config file: %v", err)
		return err
	}
	sc.Logger.Infof("Successfully updated k8s_bundle config file")
	return nil
}

func (sc *SPIREClient) WriteKubeconfig(e *Entry) error {
	if e.KubeConfig == "" {
		sc.Logger.Errorf("No KubeConfig provided in entry")
		return nil
	}
	kcDir := filepath.Join(e.SpireDir, "kubeconfigs")
	if _, err := os.Stat(kcDir); os.IsNotExist(err) {
		sc.Logger.Errorf("kubeconfig dir does not exist: %v", kcDir)
		return err
	}
	kcBytes, err := base64.StdEncoding.DecodeString(e.KubeConfig)

	if err != nil {
		sc.Logger.Errorf("Failed to decode KubeConfig: %v", err)
		return err
	}

	kcFile := filepath.Join(kcDir, e.Cluster+".yaml")

	if _, err := os.Stat(kcFile); err == nil {
		// Read the content to compare with the new content before overwriting
		currKcBytes, err := os.ReadFile(kcFile)
		if err != nil {
			sc.Logger.Errorf("Failed to read existing KubeConfig file: %v", err)
			return err
		}
		if base64.StdEncoding.EncodeToString(currKcBytes) == e.KubeConfig {
			// No change in KubeConfig, skipping write
			sc.Logger.Infof("No change in KubeConfig, skipping write to file: %v", kcFile)
			return nil
		}
	}

	sc.Logger.Infof("Writing KubeConfig to file: %v", kcFile)
	if err := os.WriteFile(kcFile, kcBytes, 0644); err != nil {
		sc.Logger.Errorf("Failed to write KubeConfig file: %v", err)
		return err
	}
	sc.Logger.Infof("Successfully wrote KubeConfig file: %v", kcFile)
	return nil
}
