FROM  registry.access.redhat.com/ubi9/go-toolset:1.23.9-1749636489@sha256:2a88121395084eaa575e5758b903fffb43dbf9d9586b2878e51678f63235b587 AS builder
WORKDIR /build
RUN git config --global --add safe.directory /build
COPY . .
RUN make build

FROM builder AS test

RUN make lint vet

FROM registry.access.redhat.com/ubi9:9.6-1749542372@sha256:861e833044a903f689ecfa404424494a7e387ab39cf7949c54843285d13a9774 AS downloader
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

FROM registry.access.redhat.com/ubi9-minimal:9.6-1749489516@sha256:f172b3082a3d1bbe789a1057f03883c1113243564f01cd3020e27548b911d3f8 AS prod
COPY --from=builder /build/terraform-repo-executor  /usr/bin/terraform-repo-executor
COPY --from=downloader /usr/bin/Terraform /usr/bin/Terraform
COPY LICENSE /licenses/LICENSE
COPY entrypoint.sh /usr/bin

RUN microdnf update -y && \
    microdnf install -y ca-certificates git && \
    microdnf clean all

ENTRYPOINT  [ "/usr/bin/entrypoint.sh" ]
