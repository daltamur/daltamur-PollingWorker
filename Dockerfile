# syntax=docker/dockerfile:1
FROM golang:alpine as builder
RUN apk --no-cache add tzdata
WORKDIR /home/dominic/GolandProjects/PollingWorker
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
WORKDIR /bin
COPY --from=builder /home/dominic/GolandProjects/PollingWorker/app/ .
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
ENV TZ=America/New_York
CMD ["./app"]