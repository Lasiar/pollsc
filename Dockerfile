FROM amd64/golang as builder
WORKDIR /go/src/github.com/Lasiar/pollsc
COPY ./*.go /go/src/github.com/Lasiar/pollsc/
COPY ./client/*.go /go/src/github.com/Lasiar/pollsc/client/
COPY ./server/*.go /go/src/github.com/Lasiar/pollsc/server/
COPY vk /go/src/github.com/Lasiar/pollsc/VK/
COPY ./base/*.go /go/src/github.com/Lasiar/pollsc/base/
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main

FROM amd64/alpine
RUN mkdir /app
COPY --from=builder /go/src/github.com/Lasiar/pollsc/main /app
WORKDIR /app
CMD ["/app/main"]
