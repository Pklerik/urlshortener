#!make
include .env
export $(shell sed 's/=.*//' .env)

upd_test:
	git fetch template && git checkout template/main .github

test:
	go test  -cover -coverprofile cover.tmp.out -coverpkg=./internal/... ./cmd/... ./pkg/...
	echo "-----------------------------------------------------------------------------------" 
	go tool cover -func cover.tmp.out
	echo "-----------------------------------------------------------------------------------"
	
bench: 
	go test -bench=. -benchmem -benchtime=100ms -run=^$$ ./...

pprof:
	go test -v ./internal/router -bench=. -benchmem -benchtime 10s -cpuprofile profiles/cpu.out -memprofile profiles/mem.out

pprof-mem:
	go tool pprof -http :9000 profiles/mem.out

pprof-cpu:
	go tool pprof -http :9000 profiles/cpu.out

lint:
	echo "================Go vet=================="
	go vet -vettool=$$(which statictest) ./...
	echo "============Go statickcheck============="
	staticcheck ./...
	echo "============Go myStaticLint============="
	go run ./cmd/staticlint $$(pwd)/cmd/shortener/...
	echo "==============Go Golint================="
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
	
# example make a iter=5 for run 1-5ths iteration
at: check_new build
	number=1 ; while [[ $$number -le $(iter) ]] ; do \
       	sudo ~/dev/shortenertest_v2 -test.v -binary-path=cmd/shortener/shortener -source-path=. -file-storage-path=local_storage.json -server-port=58080 -database-dsn=${DATABASE_DSN} -test.run="^TestIteration$$number$$" ; \
		((number = number + 1)) ; \
    done
	echo "DONE"
	
mock:
	mockgen -source=internal/repository/repository.go -destination=internal/repository/mocks/mock_links_repo.go -package=mocks
	mockgen -source=internal/service/service.go -destination=internal/service/mocks/mock_links_service.go -package=mocks
	mockgen -source=internal/config/config.go -destination=internal/config/mocks/mock_links_config.go -package=mocks

protogen:
	protoc --go_out=. --go-grpc_out=. \
	--go-grpc_opt=paths=source_relative --go_opt=paths=source_relative \
	--go_opt=default_api_level=API_OPAQUE \
	./api/proto/shortener.proto