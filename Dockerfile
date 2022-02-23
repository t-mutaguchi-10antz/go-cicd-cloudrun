# Stage 1: Go アプリのビルド

FROM golang:1.16 as builder
WORKDIR /workspace
COPY . .

## Go が参照するパッケージに GitHub のプライベートリポジトリが存在しても認証されるように設定
ARG GITHUB_TOKEN
RUN git config --global url."https://$GITHUB_TOKEN@github.com/".insteadOf "https://github.com/"
## Golang 製アプリのビルド
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /server -ldflags '-s -w' cmd/server/server.go


# Stage 2: アプリをセキュアなディストリビューションに移動

FROM gcr.io/distroless/static-debian10
COPY --from=builder /server /server
ENTRYPOINT ["/server"]