APP_NAME=rweave

all: linux mac win

linux: vet
	GOARCH=amd64 GOOS=linux go build -o ./bin/${APP_NAME}-linux-amd64 ./src/${APP_NAME}.go

mac: vet
	GOARCH=amd64 GOOS=darwin go build -o ./bin/${APP_NAME}-darwin-amd64 ./src/${APP_NAME}.go

win: vet
	GOARCH=amd64 GOOS=windows go build -x -o ./bin/${APP_NAME}-windows-windows ./src/${APP_NAME}.go

clean:
	rm -rf ./bin/*

lint:
	golint ./src

vet:
	go vet ./src

