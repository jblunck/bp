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
