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



