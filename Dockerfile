FROM golang:1.11.2-alpine3.8
RUN apk add --update make git
COPY . /go/src/github.com/ryanrolds/pedantic_orderliness
WORKDIR /go/src/github.com/ryanrolds/pedantic_orderliness
RUN make install
RUN make build
CMD ./pedantic_orderliness
