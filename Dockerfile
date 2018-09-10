FROM golang:1.10

WORKDIR /go

ENV GOBIN=/go/bin
ENV SRC_DIR=/go/src/github.com/bdelliott/dsm/

# install deps
RUN go get -v github.com/nu7hatch/gouuid
RUN go get -v go.etcd.io/etcd/clientv3

ADD . $SRC_DIR
RUN cd $SRC_DIR; go install -v $SRC_DIR/cmd/dsm.go

CMD ["/go/bin/dsm"]
