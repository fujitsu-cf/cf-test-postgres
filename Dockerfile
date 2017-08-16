FROM golang:1.7.3 as builder
WORKDIR /go/src/github.com/fujitsu-cf/cf-test-postgres/
RUN go get -d -v github.com/gorilla/mux \
    && go get -d -v github.com/lib/pq \
    && go get -d -v github.com/rs/cors 
COPY main.go    .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/fujitsu-cf/cf-test-postgres/app .
CMD ["./app"]  
