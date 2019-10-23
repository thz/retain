FROM golang as builder
MAINTAINER th-docker@thzn.de

ADD qemu-statics /usr/bin/

COPY . $GOPATH/src/github.com/thz/retain
WORKDIR $GOPATH/src/github.com/thz/retain

ARG CGO_ENABLED=0
ARG GOARCH
ARG GOARM

ENV GO111MODULE=on

RUN go version && \
	env CGO_ENABLED=${CGO_ENABLED} GOARCH=${GOARCH} GOARM=${GOARM} \
		go build -o /go/bin/retain

FROM scratch
COPY --from=builder /go/bin/retain /bin/retain
ENTRYPOINT ["/bin/retain"]
CMD []
