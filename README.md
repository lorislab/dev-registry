# dev-registry

Docker registry for local development that automatically pushes local build docker images when they are pulled out of the registry.

Example for the k3d local cluster
```
docker run -d -p 5000:5000 \
    --name k3d-registry.localhost \
    -v /var/run/docker.sock:/var/run/docker.sock \
    ghcr.io/lorislab/dev-registry:latest
```
