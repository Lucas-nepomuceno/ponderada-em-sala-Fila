FROM golang:1.25

WORKDIR /usr/src/app/backend

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -v -o /usr/local/bin/backend .

CMD ["backend"]
