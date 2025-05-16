FROM  registry.access.redhat.com/ubi9/go-toolset:1.23.6-1747333074@sha256:e0ad156b08e0b50ad509d79513e13e8a31f2812c66e9c48c98cea53420ec2bca AS builder
WORKDIR /build
RUN git config --global --add safe.directory /build
COPY . .
RUN make build

FROM builder AS test

RUN make lint vet

FROM registry.access.redhat.com/ubi9:9.6-1747219013@sha256:304b50df1ea4db9706d8a30f4bbf26f582936ebc80c7e075c72ff2af99292a54 AS downloader
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

FROM registry.access.redhat.com/ubi9-minimal:9.6-1747218906@sha256:92b1d5747a93608b6adb64dfd54515c3c5a360802db4706765ff3d8470df6290 AS prod
COPY --from=builder /build/terraform-repo-executor  /usr/bin/terraform-repo-executor
COPY --from=downloader /usr/bin/Terraform /usr/bin/Terraform
COPY LICENSE /licenses/LICENSE
COPY entrypoint.sh /usr/bin

RUN microdnf update -y && \
    microdnf install -y ca-certificates git && \
    microdnf clean all

ENTRYPOINT  [ "/usr/bin/entrypoint.sh" ]
