GOBIN    ?= ${GOPATH}/bin
IMG      ?= quay.io/konveyor/tackle2-addon-analyzer:latest
CMD      ?= bin/addon
AddonDir ?= /tmp/addon

${CMD}:
	go fmt ./...
	go vet ./...
	go build -ldflags="-w -s" -o $@ github.com/konveyor/tackle2-addon-analyzer/cmd

image-docker:
	docker build -t ${IMG} .

image-podman:
	podman build -t ${IMG} .

run: ${CMD}
	mkdir -p ${AddonDir}
	$(eval cmd := $(abspath ${CMD}))
	cd ${AddonDir};${cmd}

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -f ${CMD}
