package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	grpc "spire-api/spire-grpc"
)

const (
	AgentNamespace      = "spire"
	AgentServiceAccount = "spire-agent"
)

func Start(s string, p int, ap int, sd string, td string, uds string) {
	logger := logrus.New()
	logger.Info("Initialize api serverAndPort...")
	serverAndPort := fmt.Sprintf("%s:%d", s, p)

	spireClient, err := grpc.NewSpireClient(serverAndPort, td, uds)
	if err != nil {
		logger.Errorf("Failed to connect to SPIRE serverAndPort: %v", err)
		return
	}
	defer spireClient.GRPCConn.Close()
	router := gin.Default()
	router.GET("/v1/entries", GetEntries(spireClient))
	router.POST("/v1/entries/add", CreateEntry(spireClient, sd))
	router.POST("/v1/entries/delete", DeleteEntry(spireClient, sd))

	if err := router.Run(fmt.Sprintf(":%d", ap)); err != nil {
		logger.Errorf("Failed to start serverAndPort: %v", err)
		return
	}
}

func GetEntries(sc *grpc.SPIREClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		entries, err := sc.GetEntries()
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, entries)
	}
}

func CreateEntry(sc *grpc.SPIREClient, sd string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var e *grpc.Entry
		if err := c.ShouldBindJSON(&e); err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		e.SpireDir = sd
		entryID, err := sc.CreateEntry(e)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if e.KubeConfig != "" {
			// Update PSAT cluster and Bundle configurations if KubeConfig is provided
			err = sc.WriteKubeconfig(e)
			if err != nil {
				sc.Logger.Errorf("Failed to write kubeconfig: %v", err)
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			err = sc.AddK8sPsat(e)
			if err != nil {
				sc.Logger.Errorf("Failed to update k8s_psat config: %v", err)
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Remove bundle config since we'll be using http for bundle pull
			//err = sc.AddK8sBundle(e)
			//if err != nil {
			//	sc.Logger.Errorf("Failed to update k8s_bundle config: %v", err)
			//	c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			//	return
			//}

		} else {
			sc.Logger.Warn("No KubeConfig provided in entry, skipping K8s configuration updates")
		}

		if err := sc.SigUsr1(); err != nil {
			sc.Logger.Errorf("Failed to send SIGUSR1 to SPIRE serverAndPort: %v", err)
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, gin.H{"message": "Entry created", "entryID": entryID})
	}
}

func DeleteEntry(sc *grpc.SPIREClient, sd string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var e *grpc.Entry
		if err := c.ShouldBindJSON(&e); err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		e.SpireDir = sd
		err := sc.DeleteEntryBySPIFFE(e)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// If agent is being deleted, remove the associated K8s configurations
		if e.ServiceAccount == AgentServiceAccount && e.Namespace == AgentNamespace {
			if err := sc.DeleteK8sPsat(e); err != nil {
				sc.Logger.Errorf("Failed to delete k8s_psat config: %v", err)
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			// Remove bundle config since we'll be using http for bundle pull
			//if err := sc.DeleteK8sBundle(e); err != nil {
			//	sc.Logger.Errorf("Failed to delete k8s_bundle config: %v", err)
			//	c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			//	return
			//}

			if err := sc.DeleteKubeconfig(e); err != nil {
				sc.Logger.Errorf("Failed to delete kubeconfig: %v", err)
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if err := sc.SigUsr1(); err != nil {
				sc.Logger.Errorf("Failed to send SIGUSR1 to SPIRE serverAndPort: %v", err)
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		c.IndentedJSON(http.StatusOK, gin.H{"message": "Entry deleted"})
	}
}
