FROM quay.io/konveyor/windup-shim:latest as shim

FROM registry.access.redhat.com/ubi9/go-toolset:latest as addon
ENV GOPATH=$APP_ROOT
COPY --chown=1001:0 . .
RUN make cmd

FROM quay.io/konveyor/analyzer-lsp:latest
USER root
RUN echo -e "[centos9]" \
 "\nname = centos9" \
 "\nbaseurl = http://mirror.stream.centos.org/9-stream/AppStream/\$basearch/os/" \
 "\nenabled = 1" \
 "\ngpgcheck = 0" > /etc/yum.repos.d/centos.repo
RUN microdnf -y install \
 openssh-clients \
 subversion \
 git \
 tar
ENV HOME=/addon ADDON=/addon
WORKDIR /addon
ARG GOPATH=/opt/app-root
COPY --from=shim /usr/bin/windup-shim /usr/bin
COPY --from=addon $GOPATH/src/bin/addon /usr/bin
ENTRYPOINT ["/usr/bin/addon"]
