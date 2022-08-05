FROM golang:1.18 as builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o slack-webhook-to-telegram .

FROM alpine:3.16

WORKDIR /bin

COPY --from=builder --chown=1001:1001 /build/slack-webhook-to-telegram /bin/slack-webhook-to-telegram

USER 1001

ENTRYPOINT [ "/bin/slack-webhook-to-telegram" ]
