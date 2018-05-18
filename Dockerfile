FROM golang:1.9

ENV chatHome = /go/src/chatapp/server

RUN mkdir -p $chatHome/protos && \
    mkdir -p $chatHome/protoc

COPY ./protos/server.proto /go/src/chatapp/server/protos
COPY ./protoc /go/src/chatapp/server/protoc
COPY ./main.go /go/src/chatapp/server

WORKDIR /go/src/chatapp/server

RUN go get github.com/golang/protobuf/proto && \
    go get github.com/golang/protobuf/protoc-gen-go && \
    go get google.golang.org/grpc && \
    /go/src/chatapp/server/protoc/bin/protoc --go_out=plugins=grpc:/go/src ./protos/server.proto && \
    go install -v

CMD ["server"]

