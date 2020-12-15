BINARIES = fotolife

lint:
	golangci-lint run

test:
	go test -v

build: $(BINARIES)

fotolife: cmd/fotolife/* internal/cmd/*
	go build -v -o $@ cmd/fotolife/main.go
