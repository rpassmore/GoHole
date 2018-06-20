FROM resin/raspberrypi3-golang:latest as go-builder

#RUN [ "cross-build-start" ]
#WORKDIR "/root/gocode/src/GoHole"
WORKDIR /app
# Copy GoHole code
COPY . .
ENV GOPATH="/app"
#Install deps
RUN sh ./install.sh
# Compile
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gohole .
#RUN CGO_ENABLED=0 GOOS=linux go build -o gohole .
#RUN [ "cross-build-end" ]

###################################
#
FROM arm32v6/redis:alpine

WORKDIR /root/
COPY --from=go-builder /app/gohole .
COPY blacklists .
COPY grafana .
COPY config_example.json ./config.json
COPY docker/init.sh .

EXPOSE 24 24/udp
EXPOSE 53 53/udp
EXPOSE 443 443/udp
#ENTRYPOINT ["./init.sh"]
