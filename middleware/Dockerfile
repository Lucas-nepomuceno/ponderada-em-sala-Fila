FROM golang:1.25

WORKDIR /usr/src/app/middleware

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -v -o /usr/local/bin/middleware .

CMD ["middleware"]
