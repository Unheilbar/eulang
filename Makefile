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
#	go run cmd/eule/main.go examples/hello.eul main
#	go run cmd/eule/main.go examples/condition.eul main
#	go run cmd/eule/main.go examples/while.eul entry
#	@echo
#	go run cmd/eule/main.go examples/fcall.eul entry
#	go run cmd/eule/main.go examples/write.eul entry
#	go run cmd/eule/main.go examples/local.eul entry
#	go run cmd/eule/main.go examples/bytes32.eul entry
#	go run cmd/eule/main.go examples/precedence.eul entry
#	go run cmd/eule/main.go examples/mapping.eul entry
#	go run cmd/eule/main.go examples/params_ext.eul entry 50 true 0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed  0xa080337ae51c4e064c189e113edd0ba391df9206e2f49db658bb32cf2911730b
	go run cmd/eule/main.go examples/multi_assign.eul entry

run-bench:
	go test -v cmd/benches/*.go -bench=. -benchtime=10s -benchmem

test-compiler:
	go test -v compiler/*
