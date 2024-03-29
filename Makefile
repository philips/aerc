.POSIX:
.SUFFIXES:
.SUFFIXES: .1 .5 .7 .1.scd .5.scd .7.scd

VERSION?=0.1.1

VPATH=doc
PREFIX?=/usr/local
_INSTDIR=$(DESTDIR)$(PREFIX)
BINDIR?=$(_INSTDIR)/bin
SHAREDIR?=$(_INSTDIR)/share/aerc
MANDIR?=$(_INSTDIR)/share/man
GOFLAGS?=

GOSRC!=find . -name '*.go'
GOSRC+=go.mod go.sum

aerc: $(GOSRC)
	go build $(GOFLAGS) \
		-ldflags "-X main.Prefix=$(PREFIX) \
		-X main.ShareDir=$(SHAREDIR) \
		-X main.Version=$(VERSION)" \
		-o $@

aerc.conf: config/aerc.conf.in
	sed -e 's:@SHAREDIR@:$(SHAREDIR):g' > $@ < config/aerc.conf.in

DOCS := \
	aerc.1 \
	aerc-config.5 \
	aerc-imap.5 \
	aerc-smtp.5 \
	aerc-tutorial.7

.1.scd.1:
	scdoc < $< > $@

.5.scd.5:
	scdoc < $< > $@

.7.scd.7:
	scdoc < $< > $@

doc: $(DOCS)

all: aerc aerc.conf doc

# Exists in GNUMake but not in NetBSD make and others.
RM?=rm -f

clean:
	$(RM) $(DOCS) aerc.conf aerc

install: all
	mkdir -p $(BINDIR) $(MANDIR)/man1 $(MANDIR)/man5 $(MANDIR)/man7 \
		$(SHAREDIR) $(SHAREDIR)/filters
	install -m755 aerc $(BINDIR)/aerc
	install -m644 aerc.1 $(MANDIR)/man1/aerc.1
	install -m644 aerc-config.5 $(MANDIR)/man5/aerc-config.5
	install -m644 aerc-imap.5 $(MANDIR)/man5/aerc-imap.5
	install -m644 aerc-smtp.5 $(MANDIR)/man5/aerc-smtp.5
	install -m644 aerc-tutorial.7 $(MANDIR)/man7/aerc-tutorial.7
	install -m644 config/accounts.conf $(SHAREDIR)/accounts.conf
	install -m644 aerc.conf $(SHAREDIR)/aerc.conf
	install -m644 config/binds.conf $(SHAREDIR)/binds.conf
	install -m755 filters/hldiff $(SHAREDIR)/filters/hldiff
	install -m755 filters/html $(SHAREDIR)/filters/html
	install -m755 filters/plaintext $(SHAREDIR)/filters/plaintext

RMDIR_IF_EMPTY:=sh -c '\
if test -d $$0 && ! ls -1qA $$0 | grep -q . ; then \
	rmdir $$0; \
fi'

uninstall:
	$(RM) $(BINDIR)/aerc
	$(RM) $(MANDIR)/man1/aerc.1
	$(RM) $(MANDIR)/man5/aerc-config.5
	$(RM) $(MANDIR)/man5/aerc-imap.5
	$(RM) $(MANDIR)/man5/aerc-smtp.5
	$(RM) $(MANDIR)/man7/aerc-tutorial.7
	$(RM) -r $(SHAREDIR)
	${RMDIR_IF_EMPTY} $(BINDIR)
	$(RMDIR_IF_EMPTY) $(MANDIR)/man1
	$(RMDIR_IF_EMPTY) $(MANDIR)/man5
	$(RMDIR_IF_EMPTY) $(MANDIR)/man7
	$(RMDIR_IF_EMPTY) $(MANDIR)

.DEFAULT_GOAL := all

.PHONY: all doc clean install uninstall
