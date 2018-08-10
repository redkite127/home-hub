# BUILDER
FROM golang:1.10 AS builder
ARG SERVICE=home-hub
WORKDIR /go/src/github.com/redkite1/$SERVICE
COPY . .
RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/$SERVICE

# RUNNER
FROM alpine:3.8
WORKDIR /usr/local/bin
ARG SERVICE=home-hub
COPY --from=builder /go/bin/$SERVICE .

# it does accept the variable $SERVICE
CMD ["home-hub"]