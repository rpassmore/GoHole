FROM resin/raspberrypi3-golang:latest as go-builder

RUN [ "cross-build-start" ]
#RUN mkdir /root/gocode/src/GoHole
#WORKDIR "/root/gocode/src/GoHole"
WORKDIR /app
# Copy GoHole code
COPY . .
ENV GOPATH="/app"
#Install deps
RUN sh ./install.sh
# Compile
#RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gohole .
RUN CGO_ENABLED=0 GOOS=linux go build -o gohole .
RUN [ "cross-build-end" ]

###################################
#
FROM arm32v6/redis:alpine

WORKDIR /root/
COPY --from=go-builder /app .
COPY config_example.json .
COPY docker/init.sh .

EXPOSE 53 53/udp
EXPOSE 443 443/udp

ENTRYPOINT init.sh
