build-test:
	cd cmd/test && \
	env GOOS=linux GOARCH=arm GOARM=5 go build
push-test:
	scp ./cmd/test/test pi@10.0.0.1:~/terminal
