MAIN=cmd/crasharchive/main.go
BINARY_NAME=crasharchive
BINARY_LINUX=$(BINARY_NAME)-linux

build:
	go build -o ./bin/$(BINARY_NAME) $(MAIN)

run: build
	./bin/$(BINARY_NAME)

build/linux:
	GOOS=linux go build -o ./bin/$(BINARY_LINUX)

cli/mysql:
	docker-compose exec db mysql -p -D crash_archive

defaultconfig:
	cp ./default-docker-compose.yml ./docker-compose.yml
	cp ./config/default-config.json ./config/config.json
