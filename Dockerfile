FROM golang:1 AS builder

COPY . /build
WORKDIR /build
RUN ./build.sh

FROM mlikiowa/napcat-docker:v4.3.5

RUN arch=$(arch) && \
    if [ "$arch" = "x86_64" ]; then \
        download_link="https://mirrors.edge.kernel.org/ubuntu/pool/universe/o/olm/libolm3_3.2.10~dfsg-6ubuntu1_amd64.deb"; \
    elif [ "$arch" = "aarch64" ]; then \
        download_link="https://us.ports.ubuntu.com/pool/universe/o/olm/libolm3_3.2.10~dfsg-6ubuntu1_arm64.deb"; \
    fi && \
    curl -o libolm3.deb $download_link && \
    ls -l libolm3.deb && \
    dpkg -i libolm3.deb && rm libolm3.deb

COPY --from=builder /build/matrix-octopus-napcat /usr/bin/matrix-octopus-napcat
COPY --from=builder /build/entrypoint.sh /app/entrypoint.sh

VOLUME /app/matrix-octopus-napcat

ENTRYPOINT ["bash", "entrypoint.sh"]