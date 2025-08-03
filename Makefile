upd_test:
	git fetch template && git checkout template/main .github

test:
	go test -coverpkg=./... -cover -coferprofile cpver.tmp.out ./...
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
	go build -o ./cmd/shortener/shortener ./cmd/shortener/main.go 
