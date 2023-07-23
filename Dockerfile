FROM nvidia/cuda:12.2.0-base-ubuntu22.04
LABEL maintainer='Michaël "e7d" Ferrand <michael@e7d.io>'
RUN apt-get update && \
    apt-get -y upgrade && \
    apt-get -y install golang --no-install-recommends && \
    apt-get clean && \
    rm -r /var/lib/apt/lists/*
WORKDIR /go
COPY src/app.go app.go
RUN go build -v -o bin/app app.go
EXPOSE 9202
CMD [ "./bin/app" ]
