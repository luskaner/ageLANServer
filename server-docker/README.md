## Docker

Both a [Dockerfile](Dockerfile) and a [docker-compose](compose.yml) files are
available.
You can either first build the Dockerfile and then run it like:

``docker run -p 443:443 -p 31978:31978/udp --network=host <image>``

or just use the docker-compose file directly:

``docker compose up``

*Note: There are issues on Windows regarding UDP broadcasts due to Docker running on a VM.*