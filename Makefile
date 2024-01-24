GOPATH    ?= $(HOME)/go
GOBIN     ?= $(GOPATH)/bin
IMG       ?= quay.io/konveyor/tackle2-addon-analyzer:latest
CMD       ?= bin/addon
AddonDir  ?= /tmp/addon
GOIMPORTS = $(GOBIN)/goimports

PKG = ./cmd/... \
      ./builder/...

PKGDIR = $(subst /...,,$(PKG))


cmd: fmt vet
	go build -ldflags="-w -s" -o ${CMD} github.com/konveyor/tackle2-addon-analyzer/cmd

image-docker:
	docker build -t ${IMG} .

image-podman:
	podman build -t ${IMG} .

run: cmd
	mkdir -p ${AddonDir}
	$(eval cmd := $(abspath ${CMD}))
	cd ${AddonDir};${cmd}

fmt: $(GOIMPORTS)
	$(GOIMPORTS) -w $(PKGDIR)

vet:
	go vet $(PKG)

# Ensure goimports installed.
$(GOIMPORTS):
	go install golang.org/x/tools/cmd/goimports@latest
