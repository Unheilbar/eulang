# TODO make user to choose the example to run
run-examples:
	go run cmd/eule/main.go examples/hello.eul main
	go run cmd/eule/main.go examples/condition.eul main
	go run cmd/eule/main.go examples/while.eul entry
	@echo
	go run cmd/eule/main.go examples/fcall.eul entry
	go run cmd/eule/main.go examples/write.eul entry
	go run cmd/eule/main.go examples/local.eul entry
	go run cmd/eule/main.go examples/bytes32.eul entry
	go run cmd/eule/main.go examples/precedence.eul entry
	go run cmd/eule/main.go examples/mapping.eul entry







test-compiler:
	go test -v compiler/*
