APP?=hod
RELEASE?=0.5.2
COMMIT?=$(shell git rev-parse --short HEAD)
PROJECT?=github.com/gtfierro/hod
PERSISTDIR?=/etc/hod
PORT?=47808

clean:
	rm -f ${APP}

build: clean
	go build \
		-ldflags "-s -w -X ${PROJECT}/version.Release=${RELEASE} \
						-X ${PROJECT}/version.Commit=${COMMIT}" \
						-o ${APP}
run: build
		${APP}

container: build
	cp hod container/.
	cp Brick.ttl container/.
	cp BrickFrame.ttl container/.
	cp -r server container/.
	docker build -t gtfierro/$(APP):$(RELEASE) container

push: build
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
