FROM resin/raspberrypi3-golang:latest as go-builder
#FROM resin/amd64-golang as go-builder

#RUN [ "cross-build-start" ]
ENV GOPATH="/app/go"
WORKDIR "/app/go/src/GoHole"
# Copy GoHole code
COPY . .

#Install deps
RUN sh ./install.sh
# Compile, strip debug info -ldflags="-s -w"
#RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -a -installsuffix cgo -o gohole .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gohole .
#RUN [ "cross-build-end" ]

###################################
#
#FROM arm32v6/alpine
#FROM alpine
FROM scratch

WORKDIR /root/
COPY --from=go-builder /app/go/src/GoHole/gohole .
COPY blacklists .
COPY grafana .
COPY config_example.json ./config.json

EXPOSE 53 53/udp
EXPOSE 443 443/udp
ENTRYPOINT ["/root/gohole", "-gkey", "-s", "-c", "/root/config.json", "-abl", "/root/list.txt"]