.PHONY: build-client build-server build-all docker-build docker-push clean

build-client:
	go build -o bin/mailer-client cmd/mailer/main.go

build-server:
	go build -o bin/mailer-server cmd/server/main.go

build-all: build-client build-server

docker-build:
	docker build .

docker-push:
	docker build . -t josnelihurt/mailer-go
	docker push josnelihurt/mailer-go

clean:
	rm -rf bin/