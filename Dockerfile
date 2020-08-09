FROM golang:1.14.7
WORKDIR /go/src/github.com/transnano/http-status-for-ftp/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o health-ftp -ldflags "-s -w \
-X github.com/prometheus/common/version.Version=$(git describe --tags --abbrev=0) \
-X github.com/prometheus/common/version.BuildDate=$(date +%FT%T%z) \
-X github.com/prometheus/common/version.Branch=master \
-X github.com/prometheus/common/version.Revision=$(git rev-parse --short HEAD) \
-X github.com/prometheus/common/version.BuildUser=transnano"

FROM alpine:3.12.0
RUN apk --no-cache add ca-certificates
EXPOSE 9065
COPY --from=0 /go/src/github.com/transnano/http-status-for-ftp/health-ftp /bin/health-ftp
ENTRYPOINT ["/bin/health-ftp"]
