test:
	- CGO_CFLAGS_ALLOW=.*/github.com/gtfierro/hod/turtle go test -v
cover:
	- CGO_CFLAGS_ALLOW=.*/github.com/gtfierro/hod/turtle go test -v -coverprofile=coverage.out
	go tool cover -html=coverage.out -o=coverage.html
bench:
	- CGO_CFLAGS_ALLOW=.*/github.com/gtfierro/hod/turtle go test -v -test.bench=.
bench-cpu:
	- CGO_CFLAGS_ALLOW=.*/github.com/gtfierro/hod/turtle go test -v -test.bench=. -test.run=xyzxyzxyz -cpuprofile=cpu.out
bench-mem:
	- CGO_CFLAGS_ALLOW=.*/github.com/gtfierro/hod/turtle go test -v -test.bench=. -test.run=xyzxyzxyz -memprofile=mem.out
clean:
	rm coverage.out mem.out cpu.out block.out
