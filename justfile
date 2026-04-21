binary := if os() == "windows" { "pastebin.exe" } else { "pastebin" }

set windows-shell := ["powershell.exe", "-NoProfile", "-Command"]

default:
    @just --list

build:
    go build -o {{ binary }} .

test:
    go test -v -coverprofile coverage.out ./...

cover: test
    go tool cover -html coverage.out

clean:
    {{ if os() == "windows" { "@if (Test-Path " + binary + ") { Remove-Item " + binary + " }; if (Test-Path coverage.out) { Remove-Item coverage.out }" } else { "rm -f " + binary + " coverage.out" } }}

run *args: build
    {{ if os() == "windows" { ".\\" } else { "./" } }}{{ binary }} {{ args }}
