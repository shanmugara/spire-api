package spire_grpc

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"syscall"
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
		KubeConfigFile:          filepath.Join(e.SpireDir, "kubeconfigs", kc),
	}
	return pc
}

func (sc *SPIREClient) MakeK8sBundleCluster(e *Entry) *BundleCluster {
	sc.Logger.Infof("Creating BundleCluster instance")
	kc := e.Cluster + ".yaml"
	bc := &BundleCluster{
		KubeConfigFilePath: filepath.Join(e.SpireDir, "kubeconfigs", kc),
	}
	return bc
}

func (sc *SPIREClient) AddK8sPsat(e *Entry) error {
	currentPsat, err := sc.GetK8sPsatConfig(e)
	if err != nil {
		sc.Logger.Errorf("Failed to get current k8s_psat config: %v", err)
		return err
	}
	newCluster := sc.MakePSATCluster(e)
	UpdateCluster := false // Flag to check if we need to update an existing cluster

	// Check if cluster already exists in the current configuration
	//for _, cluster := range currentPsat.Clusters {
	//	// Check if the cluster already exists in the configuration
	//	if _, Exists := cluster[e.Cluster]; Exists {
	//		sc.Logger.Infof("Cluster %s already exists in k8s_psat config, updating it", e.Cluster)
	//		UpdateCluster = true
	//		// Update the existing cluster's KubeConfigFile if needed
	//		cluster[e.Cluster] = *newCluster
	//	}
	//}

	if _, exists := currentPsat.Clusters[0][e.Cluster]; exists {
		sc.Logger.Infof("Cluster %s already exists in k8s_psat config, updating it...", e.Cluster)
		UpdateCluster = true
		currentPsat.Clusters[0][e.Cluster] = *newCluster
	}

	// Append the new cluster to the existing clusters
	if currentPsat != nil && UpdateCluster == false {
		sc.Logger.Infof("Appending new cluster %s to k8s_psat config", e.Cluster)
		currentPsat.Clusters[0][e.Cluster] = *newCluster
		//currentPsat.Clusters = append(currentPsat.Clusters, map[string]PSATCluster{e.Cluster: *newCluster})
	}
	outFile, err := json.MarshalIndent(currentPsat, "", "  ")
	if err != nil {
		sc.Logger.Errorf("Failed to marshal updated k8s_psat config: %v", err)
		return err
	}
	// Write the updated config back to file
	if err := os.WriteFile(filepath.Join(e.SpireDir, k8sPsatConfigFile), outFile, 0644); err != nil {
		sc.Logger.Errorf("Failed to write updated k8s_psat config file: %v", err)
		return err
	}
	sc.Logger.Infof("Successfully updated k8s_psat config file")
	return nil
}

func (sc *SPIREClient) AddK8sBundle(e *Entry) error {
	currentBundle, err := sc.GetK8sBundleConfig(e)
	if err != nil {
		sc.Logger.Errorf("Failed to get current k8s_bundle config: %v", err)
		return err
	}
	newCluster := sc.MakeK8sBundleCluster(e)
	// Append the new cluster to the existing clusters

	if ok := sc.BundleExists(currentBundle, newCluster); ok {
		sc.Logger.Infof("Cluster %s already exists in k8s_bundle config, skipping update", e.Cluster)
		return nil
	}

	if currentBundle != nil {
		sc.Logger.Infof("Appending new cluster %s to k8s_bundle config", e.Cluster)
		currentBundle.Clusters = append(currentBundle.Clusters, *newCluster)
	}
	outFile, err := json.MarshalIndent(currentBundle, "", "  ")
	if err != nil {
		sc.Logger.Errorf("Failed to marshal updated k8s_bundle config: %v", err)
		return err
	}
	// Write the updated config back to file
	if err := os.WriteFile(filepath.Join(e.SpireDir, k8sBundleConfigFile), outFile, 0644); err != nil {
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

func (sc *SPIREClient) DeleteKubeconfig(e *Entry) error {

	kcFile := filepath.Join(e.SpireDir, "kubeconfigs", e.Cluster+".yaml")
	if ok := sc.KubeconfigExists(e); !ok {
		sc.Logger.Infof("Kubeconfig %s does not exist, skipping deletion", kcFile)
		return nil
	}
	sc.Logger.Infof("Deleting KubeConfig: %v", kcFile)
	if err := os.Remove(kcFile); err != nil {
		sc.Logger.Errorf("Failed to delete KubeConfig file: %v", err)
		return err
	}

	return nil
}

func (sc *SPIREClient) DeleteK8sPsat(e *Entry) error {
	currentPsat, err := sc.GetK8sPsatConfig(e)
	//updatedPsat := &K8SPSATConfig{}

	if err != nil {
		sc.Logger.Errorf("Failed to get current k8s_psat config: %v", err)
		return err
	}
	if ok := sc.PSATClusterExists(e, currentPsat); !ok {
		sc.Logger.Infof("Cluster %s does not exist in k8s_psat config, skipping deletion", e.Cluster)
		return nil
	}
	// the logic to create a new slice
	//for _, cluster := range currentPsat.Clusters {
	//	if _, ok := cluster[e.Cluster]; ok {
	//		sc.Logger.Infof("Pre Deleting cluster %s from k8s_psat config slice", e.Cluster)
	//	} else {
	//		updatedPsat.Clusters = append(updatedPsat.Clusters, cluster)
	//	}
	//}
	sc.Logger.Infof("Pre Deleting k8s_psat config: %v", e.Cluster)
	delete(currentPsat.Clusters[0], e.Cluster)

	outFile, err := json.MarshalIndent(currentPsat, "", "  ")
	if err != nil {
		sc.Logger.Errorf("Failed to marshal updated k8s_psat config: %v", err)
		return err
	}
	// Write the updated config back to file
	if err := os.WriteFile(filepath.Join(e.SpireDir, k8sPsatConfigFile), outFile, 0644); err != nil {
		sc.Logger.Errorf("Failed to write updated k8s_psat config file: %v", err)
		return err
	}
	sc.Logger.Infof("Successfully updated k8s_psat config file")

	return nil
}

func (sc *SPIREClient) DeleteK8sBundle(e *Entry) error {
	currentBundle, err := sc.GetK8sBundleConfig(e)
	if err != nil {
		sc.Logger.Errorf("Failed to get current k8s_bundle config: %v", err)
		return err
	}
	updatedBundle := &K8SBundleConfig{}
	if ok := sc.BundleExists(currentBundle, sc.MakeK8sBundleCluster(e)); !ok {
		sc.Logger.Infof("Cluster %s does not exist in k8s_bundle config, skipping deletion", e.Cluster)
		return nil
	}
	for _, cluster := range currentBundle.Clusters {
		if cluster.KubeConfigFilePath == sc.MakeK8sBundleCluster(e).KubeConfigFilePath {
			sc.Logger.Infof("Pre Deleting cluster %s from k8s_bundle config slice", e.Cluster)
			continue
		} else {
			updatedBundle.Clusters = append(updatedBundle.Clusters, cluster)
		}
	}

	outFile, err := json.MarshalIndent(updatedBundle, "", "  ")
	if err != nil {
		sc.Logger.Errorf("Failed to marshal updated k8s_bundle config: %v", err)
	}
	// Write the updated config back to file
	if err := os.WriteFile(filepath.Join(e.SpireDir, k8sBundleConfigFile), outFile, 0644); err != nil {
		sc.Logger.Errorf("Failed to write updated k8s_bundle config file: %v", err)
		return err
	}

	return nil
}

func (sc *SPIREClient) SigUsr1() error {
	// Send SIGUSR1 to the SPIRE server process
	pids, err := findPIDsByName("spire-server")
	if err != nil {
		sc.Logger.Errorf("Failed to find SPIRE server process: %v", err)
		return err
	}
	if len(pids) == 0 {
		sc.Logger.Warn("No SPIRE server process found")
		return nil
	}
	for _, pid := range pids {
		proc, err := os.FindProcess(pid)
		if err != nil {
			sc.Logger.Errorf("Failed to find SPIRE server process: %v", err)
			continue
		}
		if err := proc.Signal(syscall.SIGUSR1); err != nil {
			sc.Logger.Errorf("Failed to send SIGUSR1 to SPIRE server process: %v", err)
			continue
		}
		sc.Logger.Infof("Sent SIGUSR1 to SPIRE server process with PID: %d", pid)
	}
	return nil
}

func (sc *SPIREClient) BundleExists(currBundle *K8SBundleConfig, cl *BundleCluster) bool {
	for _, cluster := range currBundle.Clusters {
		if cluster.KubeConfigFilePath == cl.KubeConfigFilePath {
			sc.Logger.Infof("Cluster %s already exists in k8s_bundle config", cl.KubeConfigFilePath)
			return true
		}
	}
	return false

}

func (sc *SPIREClient) PSATClusterExists(e *Entry, currPsat *K8SPSATConfig) bool {
	if _, exists := currPsat.Clusters[0][e.Cluster]; exists {
		return true
	}
	return false
}

func (sc *SPIREClient) KubeconfigExists(e *Entry) bool {
	kcFile := filepath.Join(e.SpireDir, "kubeconfigs", e.Cluster+".yaml")
	if _, err := os.Stat(kcFile); err == nil {
		return true
	}
	return false
}
