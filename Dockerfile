FROM golang:latest 

RUN mkdir -p /go/src/github.com/dagozba/golangsmallshop/

WORKDIR /go/src/github.com/dagozba/golangsmallshop

COPY go.mod .
COPY go.sum .
ADD cmd/server cmd/server
ADD internal internal
ADD configs configs

ENV GO111MODULE=on

RUN go mod download && go build cmd/server/main.go

EXPOSE 50051

CMD ["./main" , "-host=:50051",  "-items-path=configs/item_definitions.yaml",  "-rules-path=configs/rules.yaml"]