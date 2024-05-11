# TODO make user to choose the example to run
run-examples:
	@go run cmd/eule/main.go examples/hello.eul
	@go run cmd/eule/main.go examples/condition.eul









test-compiler:
	go test -v compiler/*
