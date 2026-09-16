FROM  registry.access.redhat.com/ubi9/go-toolset:1.26.7-1789040808@sha256:685ccca486cc2c82b0818d835abecb1aed7a396e76ba416cfc47c28067f5d365 AS builder
WORKDIR /build
RUN git config --global --add safe.directory /build
COPY . .
RUN make build

FROM builder AS test

RUN make lint vet

FROM registry.access.redhat.com/ubi9:9.8-1789461287@sha256:c5cc9c221baa8eb13093f90c31bb58c04d83f6afd14510d3496691f92566f99e AS downloader
WORKDIR /download
ENV TENV_VERSION=3.2.10

RUN curl -sfL https://github.com/tofuutils/tenv/releases/download/v${TENV_VERSION}/tenv_v${TENV_VERSION}_Linux_x86_64.tar.gz \
    -o tenv.tar.gz \
    && tar -zvxf tenv.tar.gz

ENV TFENV_ROOT=/usr/bin
ENV TFENV_BIN=/download/tenv

RUN ${TFENV_BIN} tf install 1.4.5 && \
    ${TFENV_BIN} tf install 1.4.7 && \
    ${TFENV_BIN} tf install 1.5.7 && \
    ${TFENV_BIN} tf install 1.6.6 && \
    ${TFENV_BIN} tf install 1.7.5 && \
    ${TFENV_BIN} tf install 1.8.5

FROM registry.access.redhat.com/ubi9-minimal:9.8-1789460941@sha256:e5161a7d7d99cf22e4f34b72e111211a399d956d9b0e8714da18e9c4c8151041 AS prod
COPY --from=builder /build/terraform-repo-executor  /usr/bin/terraform-repo-executor
COPY --from=downloader /usr/bin/Terraform /usr/bin/Terraform
COPY LICENSE /licenses/LICENSE
COPY entrypoint.sh /usr/bin

RUN microdnf update -y && \
    microdnf install -y ca-certificates git && \
    microdnf clean all

ENTRYPOINT  [ "/usr/bin/entrypoint.sh" ]
