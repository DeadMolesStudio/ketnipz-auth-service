FROM golang:alpine as builder

WORKDIR /src
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \ 
	go build -mod vendor -a -installsuffix cgo -ldflags="-w -s" -o session-manager-service

FROM alpine

WORKDIR /app
COPY --from=builder /src/session-manager-service .
COPY logger/logger-config.json logger/logger-config.json

VOLUME ["/var/log/dmstudio"]

ENV db_connstr ${db_connstr}
ENV db_name ${db_name}

EXPOSE 8081
CMD ["sh", "-c", "./session-manager-service -db_connstr ${db_connstr} -db_name ${db_name}"]
