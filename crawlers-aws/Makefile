build:
	dep ensure
	env GOOS=linux go build -ldflags="-s -w" -o ../bin/crawlers-aws/hitradio-oe3 hitradio-oe3/main.go
	env GOOS=linux go build -ldflags="-s -w" -o ../bin/crawlers-aws/kronehit kronehit/main.go
