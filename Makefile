APP_NAME=rweave

build: vet
	go build -v -o ./bin/${APP_NAME} ./src/${APP_NAME}.go

clean:
	rm -rf ./bin/*

lint:
	golint ./src

vet:
	go vet ./src

