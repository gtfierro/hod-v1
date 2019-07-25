APP?=hod
RELEASE?=0.7.0-dev
GOPATH?=${HOME}/go
COMMIT?=$(shell git rev-parse --short HEAD)
PROJECT?=github.com/gtfierro/hod
PERSISTDIR?=/etc/hod
PORT?=47808

build: clean generate
	CGO_CFLAGS_ALLOW=.*/github.com/gtfierro/hod/turtle go build \
		-ldflags "-s -w -X ${PROJECT}/version.Release=${RELEASE} \
						-X ${PROJECT}/version.Commit=${COMMIT}" \
						-o ${APP}
install: generate
	CGO_CFLAGS_ALLOW=.*/github.com/gtfierro/hod/turtle go install \
		-ldflags "-s -w -X ${PROJECT}/version.Release=${RELEASE} \
						-X ${PROJECT}/version.Commit=${COMMIT}"
test: generate
	- cd db && CGO_CFLAGS_ALLOW=.*/github.com/gtfierro/hod/turtle go test -v
	- cd storage && CGO_CFLAGS_ALLOW=.*/github.com/gtfierro/hod/turtle go test -v
	#- cd lang && CGO_CFLAGS_ALLOW=.*/github.com/gtfierro/hod/turtle go test -v
	- cd turtle && CGO_CFLAGS_ALLOW=.*/github.com/gtfierro/hod/turtle go test -v

generate:
	protoc -I proto -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis proto/hod.proto -I${GOPATH}/src/github.com/dan-compton --go_out=plugins=grpc:proto
	protoc -I/usr/local/include -I. -I${GOPATH}/src -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis -I${GOPATH}/src/github.com/dan-compton --grpc-gateway_out=logtostderr=true:. proto/hod.proto
	protoc -I/usr/local/include -I. -I${GOPATH}/src -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis -I${GOPATH}/src/github.com/dan-compton --swagger_out=logtostderr=true:. proto/hod.proto
	python -m grpc_tools.protoc -Iproto -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --python_out=proto --grpc_python_out=proto proto/hod.proto
	cd server && go generate
	- cd lang && go generate

clean:
	rm -f ${APP}

run: build
		${APP}

vet:
	CGO_CFLAGS_ALLOW=.*/github.com/gtfierro/hod/turtle go vet .

container: build
	cp hod container/.
	cp Brick.ttl container/.
	cp BrickFrame.ttl container/.
	cp -r server container/.
	docker build -t gtfierro/$(APP):$(RELEASE) container

push: container
	docker push gtfierro/$(APP):$(RELEASE)

containerRun: container
	docker stop $(APP):$(RELEASE) || true && docker rm $(APP):$(RELEASE) || true
	docker run --name $(APP) \
			   --mount type=bind,source=$(shell pwd)/$(PERSISTDIR),target=/etc/hod \
			   -it \
			   -p $(PORT):47808 \
			   -e BW2_AGENT=$(BW2_AGENT) -e BW2_DEFAULT_ENTITY=$(BW2_DEFAULT_ENTITY) \
			   --rm \
			   gtfierro/$(APP):$(RELEASE)
