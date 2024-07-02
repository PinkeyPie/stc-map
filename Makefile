MODULE = $(shell go list -m)

.PHONY: generate build test lint build-docker compose compose-down migrate
generate:
	go generate ./...

build:
	go build -a -o stc-map-api-server $(MODULE)/cmd/server

test:
	go clean -testcache
	go test ./... -v

lint:
	gofmt -l .

build-docker:
	docker-build -f cmd/server/Dockerfile -t simpleServer/stc-map-api-server

compose.%:
	$(eval CMD = ${subst compose.,,$(@)})
	tools/script/compose.sh $(CMD)