DOCKER_TAG ?= ghcr.io/jblunck/bp/snapshot:latest
DOCKER_BUILD_ARGS ?=

OUTPUTDIR := target

VERSION ?= $(shell build/setlocalversion)
GIT_COMMIT_SHA ?= $(shell build/setlocalversion --git-commit-sha)
LDFLAGS := -ldflags "-w -s -X main.Version=${VERSION} -X main.GitCommitSha=${GIT_COMMIT_SHA}"

$(OUTPUTDIR)/boilerplate: cmd/boilerplate/*.go
	go build ${LDFLAGS} -tags release -o $@ $^

.DEFAULT_GOAL := all
.PHONY: all lint clean
all: lint $(OUTPUTDIR)/boilerplate

lint:
	golint -set_exit_status ./...

clean:
	@rm -vfr $(OUTPUTDIR)

.PHONY: docker
docker:
	docker build -t $(DOCKER_TAG) $(DOCKER_BUILD_ARGS) \
		--build-arg "VERSION=${VERSION}" \
		--build-arg "GIT_COMMIT_SHA=${GIT_COMMIT_SHA}" \
		-f build/Dockerfile .
