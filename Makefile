test:
	go test --cover -covermode=count -coverprofile=coverage.out ./...

build:
	bash build.sh

build-simple:
	bash build-simple.sh
	echo "Build Success!"