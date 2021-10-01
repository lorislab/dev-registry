# dev-registry

Docker registry for local development

```
docker run -d -p 5000:5000 -v /var/run/docker.sock:/var/run/docker.sock dev-registry:latest
```