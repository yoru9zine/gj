SRC = $(shell find . -name \*.go)

gj: $(SRC)
	GOGC=off go build ./cmd/gj

clean:
	rm -rf ./gj

.PHONY: clean
