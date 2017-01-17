test:
	- go test -v
	rm -rf test_databases
	git checkout test_databases
cover:
	- go test -v -coverprofile=coverage.out
	rm -rf test_databases
	git checkout test_databases
	go tool cover -html=coverage.out -o=coverage.html
clean:
	rm coverage.out mem.out cpu.out block.out
