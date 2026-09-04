GOFLAGS ?= $(GOFLAGS:)
LINTER_VERSION ?= v2.12.2

VERSION ?= $(shell git describe --abbrev=0 --tags 2>/dev/null || echo "v5.13.9-pickupFirst")

ifdef BUILD_NUMBER
VERSION := $(VERSION)+$(BUILD_NUMBER)
endif

ifdef RELEASE_VERSION
ifneq ($(RELEASE_VERSION),none)
VERSION := $(RELEASE_VERSION)
endif
endif

GO_LDFLAGS = -ldflags "-X github.com/Venafi/vcert/v5.versionString=$(VERSION) -X github.com/Venafi/vcert/v5.versionBuildTimeStamp=$(shell date -u +%Y%m%d.%H%M%S 2>/dev/null || echo "manual") -s -w"

version:
	@echo "$(VERSION)"

get: gofmt
	go get $(GOFLAGS) ./...

build_quick: get
	env GOOS=linux   GOARCH=amd64 go build $(GO_LDFLAGS) -o bin/linux/vcert         ./cmd/vcert

build: get
	env GOOS=linux   GOARCH=arm64 go build $(GO_LDFLAGS) -o bin/linux/vcert_arm       ./cmd/vcert
	env GOOS=linux   GOARCH=amd64 go build $(GO_LDFLAGS) -o bin/linux/vcert           ./cmd/vcert
	env GOOS=linux   GOARCH=386   go build $(GO_LDFLAGS) -o bin/linux/vcert86         ./cmd/vcert
	env GOOS=darwin  GOARCH=amd64 go build $(GO_LDFLAGS) -o bin/darwin/vcert          ./cmd/vcert
	env GOOS=darwin  GOARCH=arm64 go build $(GO_LDFLAGS) -o bin/darwin/vcert_arm      ./cmd/vcert
	env GOOS=windows GOARCH=amd64 go build $(GO_LDFLAGS) -o bin/windows/vcert.exe     ./cmd/vcert
	env GOOS=windows GOARCH=386   go build $(GO_LDFLAGS) -o bin/windows/vcert86.exe   ./cmd/vcert
	env GOOS=windows GOARCH=arm64 go build $(GO_LDFLAGS) -o bin/windows/vcert_arm.exe ./cmd/vcert

gofmt:
	! gofmt -l . | grep -v ^vendor/ | grep .

test:
	go test -v ./pkg/playbook/app/service/...
	go test -v ./pkg/playbook/app/vcertutil/...
	go test -v ./pkg/certificate/...
	go test -v ./pkg/endpoint/...
	go test -v ./pkg/policy/...

tpp_test:
	go test -v $(GOFLAGS) ./pkg/venafi/tpp/...

cloud_test:
	go test -v $(GOFLAGS) ./pkg/venafi/cloud/...

ngts_test:
	go test -v $(GOFLAGS) ./pkg/venafi/ngts/...

cmd_test:
	go test -v $(GOFLAGS) ./cmd/vcert/...

playbook_test:
	go test -v $(GOFLAGS) ./pkg/playbook/...

collect_artifacts:
	rm -rf artifacts
	mkdir -p artifacts
	zip -j "artifacts/vcert_$(VERSION)_linux_arm.zip" "bin/linux/vcert_arm" || exit 1
	zip -j "artifacts/vcert_$(VERSION)_linux.zip" "bin/linux/vcert" || exit 1
	zip -j "artifacts/vcert_$(VERSION)_linux86.zip" "bin/linux/vcert86" || exit 1
	zip -j "artifacts/vcert_$(VERSION)_darwin.zip" "bin/darwin/vcert" || exit 1
	zip -j "artifacts/vcert_$(VERSION)_darwin_arm.zip" "bin/darwin/vcert_arm" || exit 1
	zip -j "artifacts/vcert_$(VERSION)_windows.zip" "bin/windows/vcert.exe" || exit 1
	zip -j "artifacts/vcert_$(VERSION)_windows86.zip" "bin/windows/vcert86.exe" || exit 1
	zip -j "artifacts/vcert_$(VERSION)_windows_arm.zip" "bin/windows/vcert_arm.exe" || exit 1
