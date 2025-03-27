package main

import (
	"github.com/sirupsen/logrus"
	server "spire-api/api"
)

const (
	CERT = "spire-api.crt"
	KEY  = "spire-api.key"
	CA   = "ca.crt"
)

func main() {
	logger := logrus.New()
	logger.Info("Calling Start...")
	server.Start("omegaspire01.omegaworld.net:8081")
}

//	e := &spire_grpc.Entry{
//		TrustDomain:    "wl.dev.omegaworld.net",
//		ServiceAccount: "dummy3",
//		Namespace:      "myapp",
//		Cluster:        "ambient-b",
//	}
//	//sc.GetEntryByID("b4bd7e0d-cb1d-4a93-bb8b-fe8b5314e0ae")
//	//sc.CreateEntry(e)
//	sc.GetEntryBySPIFFE(e)
//}
