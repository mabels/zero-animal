FROM alpine
#ARG BINARY_PATH
#COPY $BINARY_PATH /usr/bin/zero-animal
COPY ./zero-animal /usr/bin/zero-animal
ENTRYPOINT ["/usr/bin/zero-animal"]
