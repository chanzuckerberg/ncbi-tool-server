FROM golang:1.7
RUN mkdir -p /go/src/ncbi-tool-server
WORKDIR /go/src/ncbi-tool-server
ADD . /go/src/ncbi-tool-server
RUN go get ./...
RUN go build
EXPOSE 80
CMD ["./ncbi-tool-server"]