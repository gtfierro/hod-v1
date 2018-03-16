APP?=hod
RELEASE?=0.5.6
COMMIT?=$(shell git rev-parse --short HEAD)
PROJECT?=github.com/gtfierro/hod
PERSISTDIR?=/etc/hod
PORT?=47808

clean:
	rm -f ${APP}

build: clean
	CGO_CFLAGS_ALLOW=.*/github.com/gtfierro/hod/turtle go build \
		-ldflags "-s -w -X ${PROJECT}/version.Release=${RELEASE} \
						-X ${PROJECT}/version.Commit=${COMMIT}" \
						-o ${APP}
install:
	CGO_CFLAGS_ALLOW=.*/github.com/gtfierro/hod/turtle go install \
		-ldflags "-s -w -X ${PROJECT}/version.Release=${RELEASE} \
						-X ${PROJECT}/version.Commit=${COMMIT}"

run: build
		${APP}

container: build
	cp hod container/.
	cp Brick.ttl container/.
	cp BrickFrame.ttl container/.
	cp -r server container/.
	docker build -t gtfierro/$(APP):$(RELEASE) container
	docker build -t gtfierro/$(APP):latest container

push: container
	docker push gtfierro/$(APP):$(RELEASE)
	docker push gtfierro/$(APP):latest

containerRun: container
	docker stop $(APP):$(RELEASE) || true && docker rm $(APP):$(RELEASE) || true
	docker run --name $(APP) \
			   --mount type=bind,source=$(shell pwd)/$(PERSISTDIR),target=/etc/hod \
			   -it \
			   -p $(PORT):47808 \
			   -e BW2_AGENT=$(BW2_AGENT) -e BW2_DEFAULT_ENTITY=$(BW2_DEFAULT_ENTITY) \
			   --rm \
			   gtfierro/$(APP):$(RELEASE)
