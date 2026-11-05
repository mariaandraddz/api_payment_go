FROM golang:1.27-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /goapp

EXPOSE 9999

CMD ["/goapp"]