include ./Makefile.Common

.PHONY: build
build:
	go build -o elastic-alert -ldflags "-s -w" ./main.go

.PHONY: elastic-alert-linux
elastic-alert-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=$(TARGETARCH) GOPROXY=$(GOPROXY) make build

.PHONY: elastic-alert-darwin
elastic-alert-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=$(TARGETARCH) GOPROXY=$(GOPROXY) make build

.PHONY: build-elastic-alert-docker
build-elastic-alert-docker-multiarch:
	export DOCKER_CLI_EXPERIMENTAL=enabled ;\
	! ( docker buildx ls | grep elastic-alert-multi-platform-builder ) && docker buildx create --use --platform=linux/amd64,linux/arm64 --name elastic-alert-multi-platform-builder ;\
	docker buildx build \
    			--builder elastic-alert-multi-platform-builder \
    			--platform linux/amd64,linux/arm64 \
    			--tag $(REGISTRY)/elastic-alert:$(TAG)  \
				--tag $(REGISTRY)/elastic-alert:latest \
    			-f Dockerfile \
    			--push \
    			.
