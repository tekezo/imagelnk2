.PHONY: debug_fetch
.PHONY: debug_run

build: staticcheck
	go build -o imagelnk2 cmd/main.go

run: staticcheck build
	PORT=1314 ./imagelnk2

debug_fetch_full:
	PORT=1315 go run debug_fetch/main.go testdata/full.jsonc

debug_fetch_single:
	PORT=1315 go run debug_fetch/main.go testdata/single.jsonc

debug_run_full:
	PORT=1316 go run debug_run/main.go testdata/full.jsonc

debug_run_single:
	PORT=1316 go run debug_run/main.go testdata/single.jsonc

staticcheck:
	staticcheck ./...

go-mod-update:
	go get -u ./...
	go mod tidy

restart-service: build
	sudo systemctl restart pqrs-imagelnk2.service
