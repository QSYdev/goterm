build-test:
	cd cmd/test && \
	env GOOS=linux GOARCH=arm GOARM=5 go build
build-ble:
	cd cmd/ble && \
	env GOOS=linux GOARCH=arm GOARM=5 go build
build-benchmark:
	cd cmd/benchmark && \
	env GOOS=linux GOARCH=arm GOARM=5 go build
push-ble:
	scp ./cmd/ble/ble pi@10.0.0.1:~/terminal-ble
push-test:
	scp ./cmd/test/test pi@10.0.0.1:~/terminal
<<<<<<< Updated upstream
protoc:
	protoc --proto_path=$(GOPATH)/src --go_out=$(GOPATH)/src $(PWD)/internal/executor/proto/executors.proto
=======
push-benchmark:
	scp ./cmd/benchmark/benchmark pi@10.0.0.1:~/terminal-benchmark
>>>>>>> Stashed changes
