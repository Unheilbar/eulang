run-examples:
	@go run cmd/eul/main.go examples/summator.easm










test-compiler:
	go test -v compiler/*
