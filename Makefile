BIN := plario

build:
	go build -o ./bin/linux/amd64/$(BIN) ./apps/cli

clean:
	find ./bin -type f -name "plario*" -exec rm {} +
