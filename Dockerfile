FROM golang:1.7
RUN mkdir -p /app
WORKDIR /app
ADD . /app
RUN go get github.com/go-sql-driver/mysql
RUN go get github.com/gorilla/mux
RUN go get github.com/mattn/go-sqlite3
RUN go build
CMD ["./app"]
