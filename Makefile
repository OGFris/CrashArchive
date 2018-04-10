MAIN=cmd/crasharchive/main.go
BINARY_NAME=crasharchive
BINARY_LINUX=$(BINARY_NAME)-linux

# build binary for native platform
build:
	go build -o ./bin/$(BINARY_NAME) $(MAIN)

# run binary
run: build
	cd bin; ./$(BINARY_NAME)

# build binary for linux
build/linux:
	GOOS=linux go build -o ./bin/$(BINARY_LINUX) $(MAIN)

# connect to the database
cli/mysql:
	docker-compose exec db mysql -p -D crash_archive

# build the containers with docker-compose
docker/build:
	docker-compose build

# run with docker as daemon
docker/run:
	docker-compose up -d

# stop the containers
docker/stop:
	docker-compose down

# remove docker images and volumes
docker/clean:
	docker-compose down -v --rmi local

# start test database and load sql from test-data
testdb/start:
	docker run --name crasharchive-test \
	--rm \
	-e MYSQL_ROOT_PASSWORD=password \
	-e MYSQL_DATABASE=crash_archive \
	-d \
	-p 3306:3306 \
	-v $(PWD)/test-data/database.sql:/docker-entrypoint-initdb.d/database.sql \
	mysql:5.7

# stop test database
testdb/stop:
	docker stop crasharchive-test

# deletes bin and docker/app folders
clean:
	rm -fr ./bin

