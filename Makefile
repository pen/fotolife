lint:
	golangci-lint run

test:
	go test -v

build: $(BINARIES)

fotolife: cmd/fotolife/* internal/cmd/*
	go build -v -ldflags '-X main.version=0.0.0-localbuild' -o $@ cmd/fotolife/main.go
