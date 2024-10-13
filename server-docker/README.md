## Docker

Both a [server Dockerfile](server/Dockerfile) and [Gen-Cert Dockerfile](genCert/Dockerfile) are available. In order to
orchestrate the server, a [docker-compose](compose.yml) file is also available.

*Note: There are issues on non Linux OSes regarding UDP broadcasts due to Docker running on a VM.*
