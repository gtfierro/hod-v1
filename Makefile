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
