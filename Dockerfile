FROM golang:1.24-alpine AS builder

RUN apk add --update make git
RUN apk --no-cache add ca-certificates

COPY . /pedantic_orderliness
WORKDIR /pedantic_orderliness

RUN go install github.com/mitranim/gow@latest
RUN go build

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app/

COPY --from=builder /go/bin/gow .
COPY --from=builder /pedantic_orderliness/pedantic_orderliness .
COPY --from=builder /pedantic_orderliness/content content

CMD ["./pedantic_orderliness"]
