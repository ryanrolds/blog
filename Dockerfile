FROM golang:1.11.2-alpine3.8
RUN apk add --update make git
COPY . /go/src/github.com/ryanrolds/pedantic_orderliness
WORKDIR /go/src/github.com/ryanrolds/pedantic_orderliness
RUN make install
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app/
COPY --from=0 /go/src/github.com/ryanrolds/pedantic_orderliness/pedantic_orderliness .
COPY --from=0 /go/src/github.com/ryanrolds/pedantic_orderliness/content content

CMD ["./pedantic_orderliness"]
