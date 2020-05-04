FROM golang:1.14-alpine

RUN apk add --update make git
RUN apk --no-cache add ca-certificates

COPY . /pedantic_orderliness
WORKDIR /pedantic_orderliness

RUN go build

#FROM alpine:latest
#RUN apk --no-cache add ca-certificates
#WORKDIR /app/
#COPY --from=0 /pedantic_orderliness .
#COPY --from=0 /pedantic_orderliness/content content

CMD ["./pedantic_orderliness"]
