BIN_PATH=./output/bin/
CMD_PATH=./cmd/
COVERAGE_PATH=./output/coverage/
COMPLEXITY_PATH=./output/complexity/

go-all: go-build go-test go-bench go-lint go-cyclo go-coverage

go-clean-build: go-clean go-nuke
	go build -o ./output/bin/ ./cmd/...

go-run-base-agent:
	go run ${CMD_PATH}base_agent

go-build:
	go build -o ./output/bin/ ./cmd/...

go-test:
	go test -v ./...

go-bench:
	go test ./... -bench=. -v

go-lint:
	go vet -v ./...
	golangci-lint -v run ./...

go-cyclo:
	gocyclo -ignore="_test.go" . > ${COMPLEXITY_PATH}complexity.txt
	cat ${COMPLEXITY_PATH}complexity.txt

go-coverage:
	go test ./... -coverprofile=${COVERAGE_PATH}coverage.out
	go tool cover -html ${COVERAGE_PATH}coverage.out -o ${COVERAGE_PATH}coverage.html

go-nuke:
	go clean -cache
	go clean -i ./...

go-clean:
	go clean
	rm ${BIN_PATH}* || true
	rm ${COVERAGE_PATH}* || true
	rm ${COMPLEXITY_PATH}* || true

# docker-server-build-run-local:
# 	docker build --platform linux/arm64  -t omnibus-test-server .
# 	docker run --platform linux/arm64 -p 8080:8080 omnibus-test-server

# docker-server-build-run-remote:
# 	echo "missing"

# docker-local-purge:
# 	- docker rm -f $(shell docker ps -aq --filter ancestor=omnibus-test-server:latest)
# 	- docker rmi -f $(shell docker images -q omnibus-test-server)

# docker-local-build-server-for-linux:
# 	docker build --platform linux/amd64  -t omnibus-test-server-linux-x64 .
# 	docker save -o output/docker_img/docker-img-omnibus-test-server-linux-x64.tar omnibus-test-server-linux-x64
# 	#docker load -i file.tar

# docker-remote-deploy-server-on-linux:
# 	docker save omnibus-test-server-linux-x64:latest | ssh dawid@192.168.1.17 'docker load && docker rm -f omnibus-test-server-linux-x64 2>/dev/null || true && docker run -d --name omnibus-test-server-linux-x64 -p 8080:8080 omnibus-test-server-linux-x64:latest'
