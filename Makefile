APP?=hod
RELEASE?=0.5.2
COMMIT?=$(shell git rev-parse --short HEAD)
PROJECT?=github.com/gtfierro/hod

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
	cp hodconfig.yaml container/.
	docker build -t gtfierro/$(APP):$(RELEASE) container

push: build
	docker push gtfierro/$(APP):$(RELEASE)

containerRun: container
	docker stop $(APP):$(RELEASE) || true && docker rm $(APP):$(RELEASE) || true
	docker run --name $(APP) -p 80:80 --rm gtfierro/$(APP):$(RELEASE)
