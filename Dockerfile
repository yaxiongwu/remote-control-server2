FROM golang:stretch

WORKDIR $GOPATH/src/github.com/YaxiongWu/remote-control-sfu

COPY go.mod go.sum ./
RUN cd $GOPATH/src/github.com/YaxiongWu/remote-control-sfu && go mod download

COPY sfu/ $GOPATH/src/github.com/YaxiongWu/remote-control-sfu/pkg
COPY cmd/ $GOPATH/src/github.com/YaxiongWu/remote-control-sfu/cmd
COPY config.toml $GOPATH/src/github.com/YaxiongWu/remote-control-sfu/config.toml
