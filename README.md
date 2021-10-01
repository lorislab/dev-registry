# dev-registry

Docker registry for local development that automatically pushes local build docker images when they are pulled out of the registry.

Example for the k3d local cluster
```
docker run -d -p 5000:5000 \
    --name k3d-registry.localhost \
    -v /var/run/docker.sock:/var/run/docker.sock \
    --label-file ./k3d-registry-labels.txt \
    ghcr.io/lorislab/dev-registry:latest
```
k3d labels `k3d-registry-labels.txt`
```
app=k3d
k3d.cluster=
k3d.registry.host=
k3d.registry.hostIP=0.0.0.0
k3d.role=registry
k3d.version=v4.4.8
k3s.registry.port.external=5000
k3s.registry.port.internal=5000
```
