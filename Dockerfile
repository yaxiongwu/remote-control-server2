FROM golang:stretch

WORKDIR $GOPATH/src/github.com/yaxiongwu/remote-control-sfu

COPY go.mod go.sum ./
RUN cd $GOPATH/src/github.com/yaxiongwu/remote-control-sfu && go mod download

COPY sfu/ $GOPATH/src/github.com/yaxiongwu/remote-control-sfu/pkg
COPY cmd/ $GOPATH/src/github.com/yaxiongwu/remote-control-sfu/cmd
COPY config.toml $GOPATH/src/github.com/yaxiongwu/remote-control-sfu/config.toml
