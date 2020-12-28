OUTPUTDIR := target

$(OUTPUTDIR)/boilerplate: cmd/boilerplate/*.go
	go build -tags release -o $@ $^

.DEFAULT_GOAL := all
.PHONY: all lint clean
all: lint $(OUTPUTDIR)/boilerplate

lint:
	golint -set_exit_status ./...

clean:
	@rm -vfr $(OUTPUTDIR)
