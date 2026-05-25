CC      	= /usr/local/go/bin/go
#CPPFLAGS # preprocessor flags
#LDFLAGS
#PACKAGE		= multiplayer
prefix		= .
exec_prefix = ${prefix}
bindir		= $(exec_prefix)/bin

.PHONY: build build-client build-server audit check clean dist distcheck distclean install install-client install-server install-strip installcheck lint test uninstall

build: build-client build-server

build-client:
	$(CC) build -o $(bindir)/client ./cmd/client/

build-server:
	$(CC) build -o $(bindir)/server ./cmd/server/

audit:
	$(CC) mod tidy -diff
	$(CC) mod verify
	test -z "$(shell gofmt -l .)"
	$(CC) vet ./...
	$(CC) run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	$(CC) run golang.org/x/vuln/cmd/govulncheck@latest ./...

check:
	echo "run the test suite, if any"

clean:
	rm -f ${bindir}/client
	rm -f ${bindir}/server

dist:
	echo "recreate package-version.tar.gz from all the source files"

#distcheck:
	# https://www.gnu.org/software/automake/manual/html_node/Preparing-Distributions.html
	# https://www.gnu.org/software/automake/manual/html_node/Checking-the-Distribution.html

# Erase anything created by `./configure`.
distclean: clean

# https://www.gnu.org/software/make/manual/html_node/DESTDIR.html
#
# This target can be thought of as a shorthand for
# `make install-exec install-data`.
# The former are architecture-dependent files and the latter are
# architecture-independent files.
# https://www.gnu.org/software/automake/manual/html_node/Two_002dPart-Install.html
install: install-client install-server

install-client:
	$(CC) install ./cmd/client/

install-server:
	$(CC) install ./cmd/server/

install-strip:
	echo "same as \`make install\` then strip debugging symbols"

installcheck:
	echo "check the installed programs or libraries, if supported"

lint:
	$(shell command -v golangci-lint) run ./...

test:
	$(CC) test -v -race -buildvcs ./...

uninstall:
	rm -f $(shell go env GOPATH)/bin/client
	rm -f $(shell go env GOPATH)/bin/server

# ---
#
# Common settings can go in a `config.site` file.
# I.e., prefix/share/config.site will define variables in `config.site` for
# ./configure --prefix ~/usr
# https://www.gnu.org/software/automake/manual/html_node/config_002esite.html
#
# VPATH builds are a way to build packages from a read-only medium or directory.
# https://www.gnu.org/software/automake/manual/html_node/VPATH-Builds.html
#

