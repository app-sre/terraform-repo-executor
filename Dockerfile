FROM  registry.access.redhat.com/ubi9/go-toolset:1.26.3-1780490420@sha256:d36470d5258da00f618b7aca9bdaab8e05134aa938bd6c42d9bd17d50ed45e76 AS builder
WORKDIR /build
RUN git config --global --add safe.directory /build
COPY . .
RUN make build

FROM builder AS test

RUN make lint vet

FROM registry.access.redhat.com/ubi9:9.8-1780554162@sha256:157ac93fde9596b9e04908fa9a22746350f53452718af9562289239f68a505c6 AS downloader
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

FROM registry.access.redhat.com/ubi9-minimal:9.8-1780378819@sha256:ae09ecc3d754bc1726cbda3e2599cc7839e09fe1cc547ce173cf669b645be3cc AS prod
COPY --from=builder /build/terraform-repo-executor  /usr/bin/terraform-repo-executor
COPY --from=downloader /usr/bin/Terraform /usr/bin/Terraform
COPY LICENSE /licenses/LICENSE
COPY entrypoint.sh /usr/bin

RUN microdnf update -y && \
    microdnf install -y ca-certificates git && \
    microdnf clean all

ENTRYPOINT  [ "/usr/bin/entrypoint.sh" ]
