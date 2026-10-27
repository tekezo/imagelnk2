.PHONY: build run test \
	debug_fetch_full debug_fetch_single debug_fetch_xcom \
	debug_run_full debug_run_single debug_run_xcom \
	staticcheck go-mod-update restart-service

build: staticcheck
	go build -o imagelnk2 cmd/main.go
	go build -o imagelnk2-cli ./cmd/imagelnk2-cli

run: staticcheck build
	PORT=1314 ./imagelnk2

test:
	go test ./...

debug_fetch_full:
	PORT=1315 go run debug_fetch/main.go testdata/full.jsonc

debug_fetch_single:
	PORT=1315 go run debug_fetch/main.go testdata/single.jsonc

debug_fetch_xcom:
	PORT=1315 go run debug_fetch/main.go testdata/xcom.jsonc

debug_run_full:
	PORT=1316 go run debug_run/main.go testdata/full.jsonc

debug_run_single:
	PORT=1316 go run debug_run/main.go testdata/single.jsonc

debug_run_xcom:
	PORT=1316 go run debug_run/main.go testdata/xcom.jsonc

staticcheck:
	staticcheck ./...

go-mod-update:
	go get -u ./...
	go mod tidy

restart-service: build
	sudo systemctl restart pqrs-imagelnk2.service
