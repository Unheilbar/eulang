run-examples:
#	@go run cmd/eule/main.go examples/hello.eul
	@go run cmd/eule/main.go examples/condition.eul









test-compiler:
	go test -v compiler/*
