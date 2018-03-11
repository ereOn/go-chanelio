build:
	golint .
	go vet .
	go test --coverprofile coverage .
	go tool cover -func=coverage

benchmark:
	go test -bench=.
