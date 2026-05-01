set quiet

default:
    just --list

build:
    go build -o formcha .

test:
    go test ./...

clean:
    rm -f formcha
