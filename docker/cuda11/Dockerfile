FROM golang:1-bullseye AS build
WORKDIR /app
COPY src/app.go app.go
RUN go build -v -o nvidiasmi-exporter app.go
RUN chmod +x nvidiasmi-exporter

FROM nvidia/cuda:11.1.1-base-ubuntu20.04
LABEL maintainer='Michaël "e7d" Ferrand <michael@e7d.io>'
COPY --from=build /app/nvidiasmi-exporter /usr/bin/nvidiasmi-exporter
RUN apt-get update && \
    apt-get -y upgrade && \
    apt-get -y autoremove --purge && \
    apt-get clean && \
    rm -r /var/lib/apt/lists/*
EXPOSE 9202
CMD ["/usr/bin/nvidiasmi-exporter"]
