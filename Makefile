.ONESHELL:
.PHONY: help install mock check build endpoints realize grpcui

define HELP
Usage:
	make [command]

Commands:
	help      ヘルプを表示
	install   開発に使用するツールをインストール
	check     フォーマット, 静的解析, ユニットテストを実行
	build     go build を実行, 実行ファイルは build/server に生成
	air       ホットリロードにより開発を補助する air を実行
endef
export HELP

define GO_TOOLS
github.com/cosmtrek/air@v1.27.10
github.com/golangci/golangci-lint/cmd/golangci-lint@v1.42.1
endef
export GO_TOOLS

help:
	@echo "$$HELP"

install:
	@for v in $$GO_TOOLS; do go install $$v; done

check:
	go fmt ./...
	golangci-lint run
	go test -cover ./... -coverprofile=docs/cover/cover.out
	go tool cover -html=docs/cover/cover.out -o docs/cover/cover.html

build:
	go build -o ./build/server -ldflags '-s -w' ./cmd/server/server.go

air:
	air