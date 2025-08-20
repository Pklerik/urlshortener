upd_test:
	git fetch template && git checkout template/main .github

test:
	go test -coverpkg=./... -cover -coferprofile cover.tmp.out ./...
	echo "-----------------------------------------------------------------------------------"
	cat cover.tmp.out | grep -v "main.go" > cover.out
	go tool cover -func cover.out
	echo "-----------------------------------------------------------------------------------"

bench:
	go test -bench=BenchmarkExecute -benchmem -benchtime 5s -count 5
	
pprof:
	go test -bench=BenchmarkExecute -benchmem -benchtime 5s -count 5 -cpuprofile cpu.out -memprofile mem.out

pprof-mem:
	go tool pprof -http :9000 mem.out

pprof-cpu:
	go tool pprof -http :9000 cpu.out

lint:
	golangci-lint run ./...

fdl:
	filedailgment --fix ./...

godot:
	godot -w ./

run:
	go run $(pwd)/cmd/shortner/main.go

build:
	go build -o ./cmd/shortener/shortener ./cmd/shortener/.

check_new:
	echo "To Do"
	
# example make a iter=5 for run 5ths iteration
at: check_new
	sudo ~/dev/shortenertestbeta -test.v -binary-path=cmd/shortener/shortener -source-path=. -server-port=8080 -test.run="^TestIteration$(iter)"