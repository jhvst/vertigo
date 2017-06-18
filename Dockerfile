# Start Debian image with latest Go version
# Workspace at /go
FROM golang

ADD . /go/src/github.com/toldjuuso/vertigo

RUN cd /go/src/github.com/toldjuuso/vertigo && go build

WORKDIR /go/src/github.com/toldjuuso/vertigo

ENTRYPOINT PORT="80" vertigo

EXPOSE 80
