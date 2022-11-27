
PREFIX?=/usr/local

BINDIR=$(PREFIX)/bin

all: tcp-hole-puncher

tcp-hole-puncher:
	go build -a -o tcp-hole-puncher

install-tcp-hole-puncher: tcp-hole-puncher
	install -Dm755 ./tcp-hole-puncher $(DESTDIR)$(BINDIR)/tcp-hole-puncher

install: install-tcp-hole-puncher

uninstall-tcp-hole-puncher:
	$(RM) $(DESTDIR)$(BINDIR)/tcp-hole-puncher

uninstall: uninstall-tcp-hole-puncher

.PHONY: all install install-tcp-hole-puncher uninstall uninstall-tcp-hole-puncher
