FROM golang:1.17-alpine

ARG VERSION
ARG COMMIT

RUN apk add make gcc libc-dev
COPY . /build
RUN cd /build && make release VERSION=$VERSION GITCOMMIT=$COMMIT

FROM alpine:latest

COPY --from=0 /build/s3-streaming-lister /usr/local/bin/s3-streaming-lister

CMD ["/usr/local/bin/s3-streaming-lister"] 
