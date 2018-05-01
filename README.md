# Docker

```bash
docker build -t "home-hub" .
docker run -d --name home-hub_01 -p 2001:2001 --restart=always home-hub
```

or using the pre-build image from Docker Hub: https://hub.docker.com/r/redkite/home-hub/
```bash
docker run -d --name home-hub_01 -p 2001:2001 --restart=always redkite/home-hub
```

# Docker Compose

```bash
docker-compose up -d
```
