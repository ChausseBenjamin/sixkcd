# sixkcd - Fetch XKCD comics from the command line
# See LICENSE file for copyright and license details.

PREFIX=/usr/local
MANPREFIX=$(PREFIX)/share/man

all:
	go mod download
	go build -o sixkcd ./sixkcd.go

install: all
	mkdir -p $(PREFIX)/bin
	mkdir -p $(MANPREFIX)/man1
	cp -f sixkcd $(PREFIX)/bin
	cp -f sixkcd.1 $(MANPREFIX)/man1

clean:
	rm -f ./sixkcd

uninstall:
	rm -f $(PREFIX)/bin/sixkcd
	rm -f $(MANPREFIX)/man1/sixkcd.1

.PHONY: all install clean uninstall help
