FROM golang:1.11.2-alpine3.8
RUN apk add --update make git
COPY . /app
WORKDIR /app
RUN make 
CMD ./pedantic_orderlieness
