
# Commands for tofu-template-tester
default:
  @just --list
# Build tofu-template-tester with Go
build:
  go build ./...

# Run tests for tofu-template-tester with Go
test:
  go clean -testcache
  go test ./...