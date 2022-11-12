.PHONY: all build-bot build-reporter test run generate lint precommit bindir format install-mockgen install-lint install-smartimports logs metrics tracing

CURDIR=$(shell pwd)
BINDIR=${CURDIR}/bin
GOVER=$(shell go version | perl -nle '/(go\d\S+)/; print $$1;')
MOCKGEN=${BINDIR}/mockgen_${GOVER}
SMARTIMPORTS=${BINDIR}/smartimports_${GOVER}
LINTVER=v1.50.0
LINTBIN=${BINDIR}/lint_${GOVER}_${LINTVER}
BOT_PACKAGE=gitlab.ozon.dev/almenschhikov/go-course-4/cmd/bot
REPORTER_PACKAGE=gitlab.ozon.dev/almenschhikov/go-course-4/cmd/reporter

ifeq ($(GOOS),)
	GOOS:=linux
endif
ifeq ($(BUILD_VERSION),)
	BUILD_VERSION:=$(shell git rev-parse --abbrev-ref HEAD)
endif
ifeq ($(BUILD_REVISION),)
	BUILD_REVISION:=$(shell git rev-parse HEAD)
endif

all: format lint build-bot build-reporter test

build-bot: TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
build-bot: LDFLAGS += -X main.version=$(BUILD_VERSION)
build-bot: LDFLAGS += -X main.gitRevision=$(BUILD_REVISION)
build-bot: LDFLAGS += -X main.buildTime=$(TIME)
build-bot: bindir
	GOOS="$(GOOS)" go build -o ${BINDIR}/bot -ldflags "$(LDFLAGS)" ${BOT_PACKAGE}

build-reporter: TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
build-reporter: LDFLAGS += -X main.version=$(BUILD_VERSION)
build-reporter: LDFLAGS += -X main.gitRevision=$(BUILD_REVISION)
build-reporter: LDFLAGS += -X main.buildTime=$(TIME)
build-reporter: bindir
	GOOS="$(GOOS)" go build -o ${BINDIR}/reporter -ldflags "$(LDFLAGS)" ${REPORTER_PACKAGE}

test:
	go test ./...

run-bot:
	go run ${BOT_PACKAGE}

run-reporter:
	go run ${REPORTER_PACKAGE}

prod-bot: build-bot
	bin/bot -c data/config.yaml 2>&1 | tee data/.logs/bot.log

prod-reporter: build-reporter
	bin/reporter -c data/config.yaml 2>&1 | tee data/.logs/reporter.log

generate: install-mockgen
	${MOCKGEN} -source=internal/clients/telegram/tgclient.go -destination=internal/mocks/clients/telegram/tgclient_mock.go
	${MOCKGEN} -source=internal/model/types.go -destination=internal/mocks/model/types_mock.go
	${MOCKGEN} -source=internal/model/currency/cbr/cbr_gateway.go -destination=internal/mocks/model/currency/cbr/cbr_gateway_mock.go
	${MOCKGEN} -source=internal/model/currency/rater.go -destination=internal/mocks/model/currency/rater_mock.go
	${MOCKGEN} -source=internal/model/expense/reporter.go -destination=internal/mocks/model/expense/reporter_mock.go
	${MOCKGEN} -source=internal/storage/types.go -destination=internal/mocks/storage/types_mock.go

lint: install-lint
	${LINTBIN} run

precommit: format lint build test
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

logs:
	mkdir -p data/.logs
	mkdir -p data/.elasticsearch
	touch data/.logs/log.txt
	touch data/.logs/offsets.yaml
	sudo chmod -R 777 data/.logs
	sudo chown -R ${UID}:${GID} data/.elasticsearch
	docker compose up graylog filed

metrics:
	mkdir -p data/.grafana
	sudo chmod -R 777 data/.grafana
	docker compose up prometheus grafana

tracing:
	docker compose up jaeger

kafka:
	docker compose up kafka
