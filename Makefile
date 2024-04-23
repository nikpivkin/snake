STATIC_DIR = ./static

build:
	go build -ldflags "-w -s" -o main main.go

build-js:
	GOOS=js GOARCH=wasm go build -ldflags "-w -s" -o $(STATIC_DIR)/main.wasm ./main_js.go


build-web:
	test -f .$(STATIC_DIR)/wasm_exec.js || cp $$(go env GOROOT)/misc/wasm/wasm_exec.js $(STATIC_DIR)/wasm_exec.js
	make build-js

serve-static: build-web
	go run ./cmd/web/main.go -a=0.0.0.0:8080 -d=$(STATIC_DIR)

.PHONY: build build-js build-web serve-static