FROM golang:1.14-alpine as builder
ENV GOPATH="$HOME/go"
RUN apk --no-cache add git gcc g++ make
WORKDIR $GOPATH/src

COPY . $GOPATH/src

RUN CGO_ENABLED=0 go get -u gopkg.in/confluentinc/confluent-kafka-go.v1/kafka
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app consumer/*.go


FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder $HOME/go/src/app .
CMD ["./app"]  