cache-benckmark:
	cd cache-benchmark && go build && mv cache-benchmark ../cache-benchmark.bak

test-http:
	go run main.go -s http & ./test.sh http

test-tcp:
	go run main.go & ./test.sh tcp