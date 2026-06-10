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

lms-set-gemma-4-12b:
	lms get gemma-4-12b-qat --yes
	lms load gemma-4-12b-qat
	lms server start --port 1234
