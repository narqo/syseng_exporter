FROM golang:1.8 AS go-builder
WORKDIR /go/src/syseng_exporter
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux make GOFLAGS="-a -installsuffix cgo"

FROM alpine:3.3
RUN apk --no-cache add ca-certificates
COPY --from=go-builder /go/src/syseng_exporter/syseng_exporter /bin/
CMD ["/bin/syseng_exporter"]
