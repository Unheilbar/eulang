# TODO make user to choose the example to run
run-examples:
#	@go run cmd/eule/main.go examples/hello.eul
#	@go run cmd/eule/main.go examples/condition.eul
#	go run cmd/eule/main.go examples/while.eul entry
#	@echo
#	go run cmd/eule/main.go examples/fcall.eul entry
#	go run cmd/eule/main.go examples/write.eul entry
#	go run cmd/eule/main.go examples/local.eul entry
#	go run cmd/eule/main.go examples/bytes32.eul entry
#	go run cmd/eule/main.go examples/precedence.eul entry
#	go run cmd/eule/main.go examples/mapping.eul entry

	go run cmd/eule/main.go examples/params_ext.eul entry






test-compiler:
	go test -v compiler/*
