# BUILDER
FROM golang:1.10 AS builder
ARG SERVICE=home-hub
COPY . /go/src/github.com/redkite1/$SERVICE
WORKDIR /go/src/github.com/redkite1/$SERVICE
RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux go install -a -installsuffix cgo  -v

# RUNNER
FROM alpine:3.7
ARG SERVICE=home-hub
WORKDIR /usr/local/bin
COPY --from=builder /go/bin/$SERVICE .

# it does accept the variable $SERVICE
CMD ["./home-hub"]