FROM nvidia/cuda:12.2.0-base-ubuntu22.04
LABEL maintainer='Michaël "e7d" Ferrand <michael@e7d.io>'
RUN apt-get update && \
    apt-get -y install git golang --no-install-recommends && \
    rm -r /var/lib/apt/lists/*
WORKDIR /go
EXPOSE 9202
CMD [ "go", "run", "app.go" ]
