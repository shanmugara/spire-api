# SPIRE Cert for Authentication

### Follow these steps to generate a SPIRE certificate for authentication

1. /opt/spire/spire-server x509 mint -dns omegaspire01.omegaworld.net -ttl 24000h -spiffeID spiffe://wl.dev.omegaworld.net/spire-api > out.txt
2. cat out.txt
3. vi api-server.crt and paste the client cert content from out.txt.
4. vi api-server.key and paste the client key content from out.txt.
5. vi ca.crt and paste the CA cert content from out.txt.
6. /opt/spire/spire-server entry create -spiffeID spiffe://wl.dev.omegaworld.net/spire-api -selector unix:uid:100 -node -admin
7. /opt/spire/spire-server entry show
8. The cets and key must be placed in a directory and mounted to the container at /certs path.

The above steps are deprectaed. We now use SPIFFE natively.

1. Install spire agent in the host.

agent.conf
```
agent {
    trust_domain = "wl.dev.omegaworld.net"
    trust_bundle_url = "https://omegaspire01.omegaworld.net/bundle.crt"
    data_dir = "/opt/spire/.data"
    log_level = "DEBUG"
    server_address = "omegaspire01.omegaworld.net"
    server_port = "8081"
    socket_path = "/tmp/spire-agent/public/api.sock"
    join_token = "06db6613-2dcd-43fe-beb4-c284230c82bd"
}

telemetry {
    Prometheus {
        port = 1234
    }
}

plugins {
    KeyManager "disk" {
        plugin_data {
            directory = "/opt/spire/.data"
        }
    }
    NodeAttestor "join_token" {
        plugin_data {

        }
    }
    WorkloadAttestor "unix" {
        plugin_data {
        }
    }
    WorkloadAttestor "docker" {
        plugin_data {
          docker_socket_path = "unix:///var/run/docker.sock"
        }
    }
}
```

2. Bundle endpoint is an nginx server.
3. Token: 
```/opt/spire/spire-server token generate --spiffeID spiffe://wl.dev.omegaworld.net/omegaspire01```
4. Register docker workload: 
```/opt/spire/spire-server entry create -parentID spiffe://wl.dev.omegaworld.net/omegaspire01 -spiffeID spiffe://wl.dev.omegaworld.net/omegaspire01/spire-api -selector docker:label:app:spire-api```
5. Run docker:
```
docker run --name spire-api -d -p 8080:8080 -v /root/gitrepos/certs:/certs/ -v /opt/spire:/opt/spire -v  /tmp/spire-agent/public/:/run/spire/sockets/ --label app=spire-api --pull always shanmugara/spire-api:v1 -api-port 8080 -port 8081 -server omegaspire01.omegaworld.net
```

