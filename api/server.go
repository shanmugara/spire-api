package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	client "spire-api/spire-grpc"
)

var sc *client.SPIREClient

func Start(s string) {
	logger := logrus.New()
	logger.Info("Initialize api server...")
	sc, err := client.NewClient(s)
	if err != nil {
		logger.Errorf("Failed to connect to SPIRE server: %v", err)
		return
	}
	defer sc.GRPCConn.Close()
	router := gin.Default()
	router.GET("/entries", GetEntries(sc))

	if err := router.Run(":8081"); err != nil {
		logger.Errorf("Failed to start server: %v", err)
		return
	}

}

func GetEntries(sc *client.SPIREClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		entries, err := sc.GetEntries()
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, entries)
	}
}
