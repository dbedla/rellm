BIN_PATH=./output/bin/
EXAMPLES_PATH=./examples/
COVERAGE_PATH=./output/coverage/
COMPLEXITY_PATH=./output/complexity/

go-all: go-build go-test go-bench go-lint go-cyclo go-coverage

go-clean-build: go-nuke go-build

go-build:
	go build -o ./output/bin/ ./examples/...
	go build -o ./output/bin/ ./e2e_tests/...

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
	go clean
	rm ${BIN_PATH}* || true
	rm ${COVERAGE_PATH}* || true
	rm ${COMPLEXITY_PATH}* || true

e2e-lms-env:
	lms unload --all
	lms get gemma-4-26b-a4b --yes
	lms load gemma-4-26b-a4b
	lms server start --port 1234

e2e-fs-tools: go-clean-build
	$(BIN_PATH)e2e_fs_agent --lms
	$(BIN_PATH)e2e_fs_agent --or-openai-luna
	$(BIN_PATH)e2e_fs_agent --or-glm53flash
	$(BIN_PATH)e2e_fs_agent --openai


e2e-format-text: go-clean-build
	$(BIN_PATH)e2e_format_text_agent --lms
	$(BIN_PATH)e2e_format_text_agent --or-openai-luna
	$(BIN_PATH)e2e_format_text_agent --or-glm53flash
	$(BIN_PATH)e2e_format_text_agent --openai

e2e-agent-switch: go-clean-build
	$(BIN_PATH)e2e_agent_switch