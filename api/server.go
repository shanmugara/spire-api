package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	client "spire-api/spire-grpc"
	grpc "spire-api/spire-grpc"
)

func Start(s string, p int, ap int) {
	logger := logrus.New()
	logger.Info("Initialize api serverAndPort...")
	serverAndPort := fmt.Sprintf("%s:%d", s, p)

	spireClient, err := client.NewClient(serverAndPort)
	if err != nil {
		logger.Errorf("Failed to connect to SPIRE serverAndPort: %v", err)
		return
	}
	defer spireClient.GRPCConn.Close()
	router := gin.Default()
	router.GET("/v1/entries", GetEntries(spireClient))
	router.POST("/v1/entries/add", CreateEntry(spireClient))
	router.POST("/v1/entries/delete", DeleteEntry(spireClient))

	if err := router.Run(fmt.Sprintf(":%d", ap)); err != nil {
		logger.Errorf("Failed to start serverAndPort: %v", err)
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

func CreateEntry(sc *client.SPIREClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var e *grpc.Entry
		if err := c.ShouldBindJSON(&e); err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err := sc.CreateEntry(e)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, gin.H{"message": "Entry created"})
	}
}

func DeleteEntry(sc *client.SPIREClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var e *grpc.Entry
		if err := c.ShouldBindJSON(&e); err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err := sc.DeleteEntryBySPIFFE(e)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, gin.H{"message": "Entry deleted"})
	}
}
