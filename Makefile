build:
	golint .
	go vet .
	go test -race -coverprofile=coverage.txt -covermode=atomic
	go tool cover -func=coverage.txt

benchmark:
	go test -bench=.
