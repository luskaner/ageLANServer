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
RUN go build -o build/bin/genCert ./server-genCert
RUN go build -o build/server ./server
RUN build/bin/genCert
RUN rm -rf build/bin
RUN rm -rf build/resources/unix

# Execute
FROM alpine:3.20
EXPOSE 443
VOLUME /app/resources
VOLUME /app/logs
WORKDIR /app
COPY --from=compiler /app/build .
ENTRYPOINT ["./server"]