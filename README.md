# go-github-sample

## Data Source

- [Issues · sataga/issue-warehouse](https://github.com/sataga/issue-warehouse/issues?q=)

## For Testing

```sh
# 開発中にテストする場合
go test -v ./...

```

```sh
# 開発中にMockを作り直す場合
go generate ./...
goimports -w .

```
