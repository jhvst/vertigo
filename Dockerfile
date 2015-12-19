#Start Debian image with latest Go version
#Workspace at /go
FROM golang

ADD . /go/src/github.com/9uuso/vertigo

RUN go get github.com/tools/godep && cd /go/src/github.com/9uuso/vertigo && godep go build

WORKDIR /go/src/github.com/9uuso/vertigo

ENTRYPOINT PORT="80" vertigo

EXPOSE 80
