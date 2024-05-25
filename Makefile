version?=1.0

APP_NAME=server
AGENT_NAME=agent
STATICLINT=staticlint
buildVersion=$(shell git log --pretty=format:"%h" -1)
buildCommit=$(shell git log --pretty=format:"%s (%ad)" --date=rfc2822 -1)
buildDate=$(shell date +'%Y-%m-%d %H:%M:%S')

dir:
	mkdir -p ./bin

clean_app: dir
	rm -f ./bin/${APP_NAME}
clean_agent: dir
	rm -f ./bin/${AGENT_NAME}
clean_staticlint: dir
	rm -f ./bin/${STATICLINT}

clean: clean_agent clean_app clean_staticlint

build_app: clean_app
	go build -ldflags "-X 'main.buildVersion=${version} (${buildVersion})' -X 'main.buildDate=${buildDate}' -X 'main.buildCommit=${buildCommit}'" -o "./bin/${APP_NAME}" ./cmd/${APP_NAME}/*.go

build_agent: clean_agent
	go build -ldflags "-X 'main.buildVersion=${version} (${buildVersion})' -X 'main.buildDate=${buildDate}' -X 'main.buildCommit=${buildCommit}'" -o "./bin/${AGENT_NAME}" ./cmd/${AGENT_NAME}/*.go

build_staticlint: clean_staticlint
	go build -o "./bin/${STATICLINT}" ./cmd/${STATICLINT}/*.go

build: build_app build_agent build_staticlint

run_app: build_app
	./bin/${APP_NAME}

run_agent: build_agent
	./bin/${AGENT_NAME}

run_staticlint: build_staticlint
	go vet -vettool=./bin/${STATICLINT} ./...

test:
	go test -v -count=1 ./...

race:
	go test -v -race -count=1 ./...

.PHONY: cover
cover:
	go test -v -coverpkg=./... -coverprofile=coverage.out -covermode=count ./...
	go tool cover -func coverage.out
