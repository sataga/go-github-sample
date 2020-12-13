FROM golang:latest

#ディレクトリ作成
WORKDIR /go/src/usersupport-dairy-report
#ホストOSのmain.goをWORKDIRにコピー
COPY . .

#バイナリを生成
RUN go install -v . && \
    CGO_ENABLE=0 GOOS=linux GOARCH=amd64 go build -o ./bin/usersupport-dairy-report .

#バイナリを実行
ENTRYPOINT ["./bin/usersupport-dairy-report"]
CMD ["us"]
