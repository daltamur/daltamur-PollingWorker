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
ARG AWS_REGION_ARG
ENV AWS_REGION=$AWS_REGION_ARG
ARG AWS_ACCESS_KEY_ID_ARG
ENV AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID_ARG
ARG AWS_SECRET_ACCESS_KEY_ARG
ENV  AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY_ARG
ARG LOGGLY_TOKEN_ARG
ENV LOGGLY_TOKEN=$LOGGLY_TOKEN_ARG
ARG LOGGLY_TOKEN_ARG
ARG LASTFM_TOKEN_ARG
ENV LASTFM_TOKEN=$LASTFM_TOKEN_ARG
ENV TZ=America/New_York
CMD ["./app"]