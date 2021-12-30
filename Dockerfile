FROM golang:1.17-alpine

ARG VERSION
ARG COMMIT

RUN apk add make gcc libc-dev
COPY . /build
RUN cd /build && make build VERSION=$VERSION GITCOMMIT=$COMMIT

FROM alpine:latest

COPY --from=0 /build/zero-animal /usr/local/bin/zero-animal

CMD ["/usr/local/bin/zero-animal"] 
