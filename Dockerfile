ARG BUILDPLATFORM=
FROM alpine@$BUILDPLATFORM
#@sha256:e7d88de73db3d3fd9b2d63aa7f447a10fd0220b7cbf39803c803f2af9ba256b3
#ARG BINARY_PATH
#COPY $BINARY_PATH /usr/bin/zero-animal
COPY ./zero-animal /usr/bin/zero-animal
ENTRYPOINT ["/usr/bin/zero-animal"]
