DIST_PATH ?= ./dist
MAIN ?= main.go
NAME ?= iplayer-to-plex

build:
	go build -o ${DIST_PATH}/${NAME}${EXT} ${MAIN}

	cd ${DIST_PATH} && sha256sum ${NAME}${EXT} > ${NAME}${EXT}.sha256
.PHONY: build

build-all:
	make clean

	EXT=-darwin-386 GOARCH=386 GOOS=darwin make build
	EXT=-darwin-amd64 GOARCH=amd64 GOOS=darwin make build

	EXT=-linux-arm GOARCH=arm GOOS=linux make build
	EXT=-linux-arm64 GOARCH=arm64 GOOS=linux make build
	EXT=-linux-amd64 GOARCH=amd64 GOOS=linux make build
	EXT=-linux-386 GOARCH=386 GOOS=linux make build

	EXT=-win-386.exe GOARCH=386 GOOS=windows make build
	EXT=-win-amd64.exe GOARCH=amd64 GOOS=windows make build
.PHONY: build-all

clean:
	rm -Rf ${DIST_PATH}
.PHONY: clean

install:
	dep ensure -v
.PHONY: install

run:
	go run ${MAIN} ${ARGS}
.PHONY: run

version:
	rm ./VERSION
	echo ${VER} > ./VERSION

	git add ./VERSION
	git commit -m "v${VER}"
	git tag "v${VER}"
	git push --tags
	git push
.PHONY: version
