# syntax=docker/dockerfile:1

# Compile
FROM golang:1.22-alpine3.20 AS compiler
WORKDIR /app
COPY server-docker/go.work.template go.work
COPY common common
COPY server server
COPY server-genCert server-genCert
RUN mkdir build
RUN cp -r server/resources build/resources
RUN go build -o build/bin/genCert ./server-genCert && \
    build/bin/genCert && \
    rm -rf build/bin
RUN go build -ldflags="-s -w" -o build/server ./server
# Compress
FROM alpine:3.20 AS compressor
RUN apk add --no-cache upx
WORKDIR /app
COPY --from=compiler /app/build/server .
RUN upx --lzma server
# Execute
FROM scratch
EXPOSE 443
VOLUME /app/resources
VOLUME /app/logs
WORKDIR /app
COPY --from=compiler /app/build/resources resources
COPY --from=compressor /app/server .
ENTRYPOINT ["./server"]