test:
	go test -race -covermode=atomic -coverprofile=coverage.out ./...

build:
	bash build.sh

build-simple:
	bash build-simple.sh
	echo "Build Success!"