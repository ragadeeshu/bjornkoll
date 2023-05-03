FROM golang:1.18-alpine
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download && go mod verify

COPY html ./html
COPY main.go ./

RUN go build

EXPOSE 8080

ENTRYPOINT ./bjornkoll