FROM  registry.access.redhat.com/ubi9/go-toolset:1.23.9-1751538372@sha256:381fb72f087a07432520fa93364f66b5981557f1dd708f3c4692d6d0a76299b3 AS builder
WORKDIR /build
RUN git config --global --add safe.directory /build
COPY . .
RUN make build

FROM builder AS test

RUN make lint vet

FROM registry.access.redhat.com/ubi9:9.6-1751445649@sha256:2e4eebec441e8bbc3459fcc83ddee0f7d3cfd219097b4110a37d7ff4fe0ff2e9 AS downloader
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

FROM registry.access.redhat.com/ubi9-minimal:9.6-1751286687@sha256:383329bf9c4f968e87e85d30ba3a5cb988a3bbde28b8e4932dcd3a025fd9c98c AS prod
COPY --from=builder /build/terraform-repo-executor  /usr/bin/terraform-repo-executor
COPY --from=downloader /usr/bin/Terraform /usr/bin/Terraform
COPY LICENSE /licenses/LICENSE
COPY entrypoint.sh /usr/bin

RUN microdnf update -y && \
    microdnf install -y ca-certificates git && \
    microdnf clean all

ENTRYPOINT  [ "/usr/bin/entrypoint.sh" ]
