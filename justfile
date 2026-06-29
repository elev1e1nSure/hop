_default:
    just --list

run language="":
    go run ./cmd/hop {{if language != "" { "--language " + language } else { "" }}}

build:
    go build -o hop.exe ./cmd/hop

test:
    go test ./...

lint:
    go vet ./...

fmt:
    gofmt -s -w . && golines -w --max-len=120 .
