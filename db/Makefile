test:
	- go test -v
	rm -rf test_databases
	git checkout test_databases
cover:
	- go test -v -coverprofile=coverage.out
	rm -rf test_databases
	git checkout test_databases
	go tool cover -html=coverage.out -o=coverage.html
bench:
	- go test -v -test.bench=BenchmarkQuery -test.run=xyzxyzxyz
	rm -rf test_databases
	git checkout test_databases
bench-cpu:
	- go test -v -test.bench=BenchmarkQuery -test.run=xyzxyzxyz -cpuprofile=cpu.out
	rm -rf test_databases
	git checkout test_databases
bench-mem:
	- go test -v -test.bench=BenchmarkQuery -test.run=xyzxyzxyz -memprofile=mem.out
	rm -rf test_databases
	git checkout test_databases
clean:
	rm coverage.out mem.out cpu.out block.out
