run:
	./run.sh

tests:
	go test ./... -v -coverpkg=./... -coverprofile=profile.cov ./...
	go tool cover -func profile.cov

PHONY: run tests