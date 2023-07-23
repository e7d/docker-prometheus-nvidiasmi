# Docker Prometheus Nvidia SMI Exporter

Dockerized Prometheus exporter for GPU statistics from [nvidia-smi](https://developer.nvidia.com/nvidia-system-management-interface), written in Go.
Supports multiple GPUs.

# Supported tags and respective `Dockerfile` links

- [`cuda12`, `latest`](docker/cuda12/Dockerfile)
- [`cuda11`](docker/cuda11/Dockerfile)

# How to use

Run with a Docker command:
```
docker run --privileged --runtime nvidia -p 9202:9202/tcp e7db/prometheus-nvidiasmi:cuda12
```

Run through a docker-compose file:
```
services:
  prometheus-nvidiasmi:
    image: e7db/prometheus-nvidiasmi:cuda12
    privileged: true
    runtime: nvidia
    ports:
      - "9202:9202/tcp"
```

Check result at: [http://localhost:9202/metrics](http://localhost:9202/metrics)

# Grafana dashboard

[Nvidia SMI Metrics dashboard](https://grafana.com/grafana/dashboards/12357) on Grafana Labs
