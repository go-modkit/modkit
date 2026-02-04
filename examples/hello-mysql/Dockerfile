FROM golang:1.25

WORKDIR /repo/examples/hello-mysql

COPY . /repo

RUN go mod download

CMD ["sh", "-c", "go run ./cmd/migrate && go run ./cmd/seed && go run ./cmd/api"]
