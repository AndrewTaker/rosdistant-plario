BIN := plario

build:
	go build -o ./bin/$(BIN) ./apps/cli

clean:
	find ./bin -type f -name "plario*" -exec rm {} +
