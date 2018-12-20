FROM golang:alpine as builder
RUN apk update && apk add git
WORKDIR /go/src/github.com/Cloud-Pie/APPA/server
COPY . .

RUN go get -u github.com/golang/dep/cmd/dep
RUN $GOPATH/bin/dep ensure -vendor-only

RUN go build -o main .
FROM alpine
RUN adduser -S -D -H -h /app appuser
USER appuser
COPY --from=builder /go/src/github.com/Cloud-Pie/APPA/server/main /app/
COPY --from=builder /go/src/github.com/Cloud-Pie/APPA/server/assets /app/assets/
COPY --from=builder /go/src/github.com/Cloud-Pie/APPA/server/templates /app/templates/
WORKDIR /app
CMD ["./main"]
