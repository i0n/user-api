SHELL := /bin/bash
NAME := user-api
CONTAINER_NAME := i0nw/${NAME}
ROOT_PACKAGE := github.com/i0n/user-api
GO := go

REV := $(shell git rev-parse --short HEAD 2> /dev/null || echo 'unknown')
GO_VERSION := $(shell $(GO) version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/')

BRANCH     := $(shell git rev-parse --abbrev-ref HEAD 2> /dev/null  || echo 'unknown')
BUILD_DATE := $(shell date +%Y%m%d-%H:%M:%S)
BUILD_USER := $(shell whoami)
CGO_ENABLED = 0

all: build

check: fmt build test

POSTGRES_USER := postgres
POSTGRES_PASSWORD := mysecretpassword
POSTGRES_URL := 0.0.0.0:5432
POSTGRES_DB := user-api

version:
ifeq (,$(wildcard pkg/version/VERSION))
TAG := $(shell git fetch --all -q 2>/dev/null && git describe --abbrev=0 --tags 2>/dev/null)
ON_EXACT_TAG := $(shell git name-rev --name-only --tags --no-undefined HEAD 2>/dev/null | sed -n 's/^\([^^~]\{1,\}\)\(\^0\)\{0,1\}$$/\1/p')
VERSION := $(shell [ -z "$(ON_EXACT_TAG)" ] && echo "$(TAG)-dev-$(REV)" | sed 's/^v//' || echo "$(TAG)" | sed 's/^v//' )
else
VERSION := $(shell cat pkg/version/VERSION)
endif
BUILDFLAGS := -ldflags \
  " -X $(ROOT_PACKAGE)/pkg/version.Version=$(VERSION)\
		-X $(ROOT_PACKAGE)/pkg/version.Revision='$(REV)'\
		-X $(ROOT_PACKAGE)/pkg/version.Branch='$(BRANCH)'\
		-X $(ROOT_PACKAGE)/pkg/version.BuildDate='$(BUILD_DATE)'\
		-X $(ROOT_PACKAGE)/pkg/version.BuildUser='$(BUILD_USER)'\
		-X $(ROOT_PACKAGE)/pkg/version.GoVersion='$(GO_VERSION)'"

DOCKER_BUILDFLAGS := -ldflags \
  " -X $(ROOT_PACKAGE)/pkg/version.Version=$(DOCKER_ARG_VERSION)\
		-X $(ROOT_PACKAGE)/pkg/version.Revision='$(DOCKER_ARG_REV)'\
		-X $(ROOT_PACKAGE)/pkg/version.Branch='$(DOCKER_ARG_BRANCH)'\
		-X $(ROOT_PACKAGE)/pkg/version.BuildDate='$(BUILD_DATE)'\
		-X $(ROOT_PACKAGE)/pkg/version.BuildUser='$(DOCKER_ARG_BUILD_USER)'\
		-X $(ROOT_PACKAGE)/pkg/version.GoVersion='$(GO_VERSION)'"

DOCKER_NETWORK := $(shell docker network ls --filter name=user-api -q)

print-version: version
	@echo $(VERSION)

print-rev:
	@echo $(REV)

print-branch:
	@echo $(BRANCH)

print-build-date:
	@echo $(BUILD_DATE)

print-build-user:
	@echo $(BUILD_USER)

print-go-version:
	@echo $(GO_VERSION)

build: version
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(BUILDFLAGS) -o build/$(NAME) main.go

linux: version
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 $(GO) build $(BUILDFLAGS) -o build/linux/$(NAME) main.go

linux-from-docker:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 $(GO) build $(DOCKER_BUILDFLAGS) -o build/linux/$(NAME) main.go

docker-create-user-api-network:
ifeq ($(strip $(DOCKER_NETWORK)),)
	@echo Creating docker network user-api...
	docker network create user-api
else
	@echo Docker network user-api already created.
endif

docker-build: print-version print-rev print-branch
	docker build --no-cache . --build-arg DOCKER_ARG_VERSION=$(VERSION) --build-arg DOCKER_ARG_REV=$(REV) --build-arg DOCKER_ARG_BRANCH=$(BRANCH) --build-arg DOCKER_ARG_BUILD_USER=${BUILD_USER} -t ${CONTAINER_NAME}:latest
	docker tag ${CONTAINER_NAME}:latest ${CONTAINER_NAME}:$(VERSION)

docker-run: docker-build docker-create-user-api-network
	docker run --name user-api --rm --network user-api -e POSTGRES_USER=${POSTGRES_USER} -e POSTGRES_PASSWORD=${POSTGRES_PASSWORD} -e POSTGRES_URL=postgres:5432 -e POSTGRES_DB=${POSTGRES_DB}	-p 8080:8080 ${CONTAINER_NAME}:latest

docker-run-ci: docker-build docker-create-user-api-network
	docker run --name user-api --rm -d --network user-api -e POSTGRES_USER=${POSTGRES_USER} -e POSTGRES_PASSWORD=${POSTGRES_PASSWORD} -e POSTGRES_URL=postgres:5432 -e POSTGRES_DB=${POSTGRES_DB}	-p 8080:8080 ${CONTAINER_NAME}:latest

docker-run-postgres-dev: docker-create-user-api-network
	docker run --name postgres --network user-api -e POSTGRES_PASSWORD=${POSTGRES_PASSWORD} -e POSTGRES_DB=${POSTGRES_DB} -v $(shell pwd)/db/:/docker-entrypoint-initdb.d/  --rm -p 5432:5432 postgres

docker-run-postgres-ci: docker-create-user-api-network
	docker run --name postgres --network user-api -e POSTGRES_PASSWORD=${POSTGRES_PASSWORD} -e POSTGRES_DB=${POSTGRES_DB} -v $(shell pwd)/db/:/docker-entrypoint-initdb.d/  --rm -d -p 5432:5432 postgres

docker-run-psql-dev:
	docker exec -it postgres  psql -U ${POSTGRES_USER} ${POSTGRES_DB}

docker-push:
	docker push ${CONTAINER_NAME}:latest
	docker push ${CONTAINER_NAME}:$(VERSION)

test-integration:
	USER_API_URL=0.0.0.0:8080 k6 run ./test/integration/k6.js

kubernetes-rolling-update-current-version:
	kubectl set image -f kube/deployment.yaml app=${CONTAINER_NAME}:${VERSION}

kubernetes-rolling-update-latest:
	kubectl set image -f kube/deployment.yaml app=${CONTAINER_NAME}:latest

deploy: clean docker-build docker-push kubernetes-rolling-update-current-version

run: 	
	POSTGRES_USER=${POSTGRES_USER} POSTGRES_PASSWORD=${POSTGRES_PASSWORD} POSTGRES_URL=${POSTGRES_URL} POSTGRES_DB=${POSTGRES_DB} go run main.go

get-test-deps:
	@$(GO) get github.com/axw/gocov/gocov
	@$(GO) get -u gopkg.in/matm/v1/gocov-html

test:
	@CGO_ENABLED=$(CGO_ENABLED) $(GO) test -count=1 -coverprofile=cover.out -failfast -short -parallel 12 ./...

test-report: get-test-deps test
	@gocov convert cover.out | gocov report

test-report-html: get-test-deps test
	@gocov convert cover.out | gocov-html > cover.html && open cover.html

clean:
	rm -rf build release cover.out cover.html dist
