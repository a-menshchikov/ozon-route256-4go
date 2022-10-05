CURDIR=$(shell pwd)
BINDIR=${CURDIR}/bin
GOVER=$(shell go version | perl -nle '/(go\d\S+)/; print $$1;')
MOCKGEN=${BINDIR}/mockgen_${GOVER}
SMARTIMPORTS=${BINDIR}/smartimports_${GOVER}
LINTVER=v1.50.0
LINTBIN=${BINDIR}/lint_${GOVER}_${LINTVER}
PACKAGE=gitlab.ozon.dev/almenschhikov/go-course-4/cmd/bot

ifeq ($(GOOS),)
	GOOS:=linux
endif
ifeq ($(BUILD_VERSION),)
	BUILD_VERSION:=$(shell git rev-parse --abbrev-ref HEAD)
endif
ifeq ($(BUILD_REVISION),)
	BUILD_REVISION:=$(shell git rev-parse HEAD)
endif

all: format build test lint

build: TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
build: LDFLAGS += -X main.version=$(BUILD_VERSION)
build: LDFLAGS += -X main.gitRevision=$(BUILD_REVISION)
build: LDFLAGS += -X main.buildTime=$(TIME)
build: bindir
	GOOS="$(GOOS)" go build -o ${BINDIR}/bot -ldflags "$(LDFLAGS)" ${PACKAGE}

test:
	go test ./...

run:
	go run ${PACKAGE}

generate: install-mockgen
	${MOCKGEN} -source=internal/model/bot.go -destination=internal/mocks/model/bot_mock.go

lint: install-lint
	${LINTBIN} run

precommit: format build test lint
	echo "OK"

bindir:
	mkdir -p ${BINDIR}

format: install-smartimports
	${SMARTIMPORTS} -exclude internal/mocks

install-mockgen: bindir
	test -f ${MOCKGEN} || \
		(GOBIN=${BINDIR} go install github.com/golang/mock/mockgen@v1.6.0 && \
		mv ${BINDIR}/mockgen ${MOCKGEN})

install-lint: bindir
	test -f ${LINTBIN} || \
		(GOBIN=${BINDIR} go install github.com/golangci/golangci-lint/cmd/golangci-lint@${LINTVER} && \
		mv ${BINDIR}/golangci-lint ${LINTBIN})

install-smartimports: bindir
	test -f ${SMARTIMPORTS} || \
		(GOBIN=${BINDIR} go install github.com/pav5000/smartimports/cmd/smartimports@latest && \
		mv ${BINDIR}/smartimports ${SMARTIMPORTS})

docker-run:
	sudo docker compose up
