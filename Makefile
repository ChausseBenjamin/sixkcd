# sixkcd - Fetch XKCD comics from the command line
# See LICENSE file for copyright and license details.
COMMIT := $(shell git rev-parse --short HEAD)
VERSION := $(or $(SIXKCD_VERSION),dev-$(COMMIT))

SRC := $(shell go list -f '{{range .GoFiles}}{{$$.Dir}}/{{.}} {{end}}' ./...)

CMD := sixkcd

PREFIX=/usr/local
MANPREFIX=$(PREFIX)/share/man

sixkcd: $(SRC)
	go mod download
	go mod verify
	go build -ldflags "-X main.version=$(VERSION)" -o $(CMD) ./$(CMD).go

install: all
	mkdir -p $(PREFIX)/bin
	mkdir -p $(MANPREFIX)/man1
	cp -f $(CMD) $(PREFIX)/bin
	cp -f $(CMD).1 $(MANPREFIX)/man1

clean:
	rm -f ./$(CMD)

uninstall:
	rm -f $(PREFIX)/bin/$(CMD)
	rm -f $(MANPREFIX)/man1/$(CMD).1

info:
	@echo "$(CMD)"
	@echo "Version: $(VERSION)"

all: info sixkcd

.PHONY: all install clean uninstall help
