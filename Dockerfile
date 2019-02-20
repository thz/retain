FROM golang:alpine as builder
RUN apk add --update git

COPY . $GOPATH/src/github.com/thz/retain
WORKDIR $GOPATH/src/github.com/thz/retain

ARG CGO_ENABLED=0
ARG GOARCH=amd64
ARG GOARM=6

ENV GO111MODULE=on

RUN go version && \
	env CGO_ENABLED=${CGO_ENABLED} GOARCH=${GOARCH} GOARM=${GOARM} \
	go build -o /go/bin/retain

FROM scratch
COPY --from=builder /go/bin/retain /bin/retain
ENTRYPOINT ["/bin/retain"]
CMD []
