FROM golang:1.10-alpine as build
RUN apk --no-cache add --virtual .build-deps\
    git make build-base && \
    go get -u github.com/golang/dep/cmd/dep
WORKDIR /go/src/ledger.api

COPY Gopkg.lock Gopkg.toml /go/src/ledger.api/
RUN dep ensure -vendor-only

COPY . /go/src/ledger.api
RUN make

FROM alpine:3.8
RUN apk --no-cache add ca-certificates
COPY --from=build /go/bin /usr/local/bin
CMD ["ledger-api"]