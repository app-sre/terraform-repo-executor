FROM  registry.access.redhat.com/ubi9/go-toolset:1.21.13-2.1729776560 AS builder
WORKDIR /build
RUN git config --global --add safe.directory /build
COPY . .
RUN make lint build

FROM registry.access.redhat.com/ubi9:9.4 AS downloader
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

FROM registry.access.redhat.com/ubi9-minimal:9.4
COPY --from=builder /build/terraform-repo-executor  /usr/bin/terraform-repo-executor
COPY --from=downloader /usr/bin/Terraform /usr/bin/Terraform

RUN microdnf update -y && \
    microdnf install -y ca-certificates && \
    microdnf clean all

ENTRYPOINT  [ "/usr/bin/terraform-repo-executor" ]
