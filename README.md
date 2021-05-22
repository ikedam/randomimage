docker-compose run --rm -e GOOS=freebsd -e GOARCH=amd64 dev go build -o random.cgi ./cmd/cgi
