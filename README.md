# go-musthave-shortener-tpl

**моки**
```
docker run -v "$PWD":/src -w /src vektra/mockery --all
```

**развитие**
- [ ] оптимизировать мидлварь DeGzip, чтоб не аллоцировать новый ридер на каждый запрос. для этого можно использовать [sync.Pool](https://pkg.go.dev/sync#Pool), но нужно лучше разбираться в примитивах синхронизации. опираться на то, как сделано в библиотеке [gziphandler](https://github.com/NYTimes/gziphandler)

**команды**

***починить импорты***
```
goimports -local "github.com/dtroode/urlshorter" -w .
```

**сборка с флагами**

в приложение можно передать информацию о версии, дате сборки и коммите, с которого собрано приложение. тогда оно будет логировать эту информацию при старте. для передачи нужно использовать [флаги линковки](https://pkg.go.dev/cmd/link). например при сборке это можно сделать так:
```
go build -ldflags "-X main.buildVersion=0.1 -X 'main.buildDate=$(date +'%Y/%m/%d %H:%M:%S')' -X 'main.buildCommit=$(git show --pretty=format:"%H" --no-patch)'" -o cmd/shortener/shortener cmd/shortener/main.go
```

тогда лог будет таким:
```
Build version: 0.1
Build date: 2025/07/01 17:49:42
Build commit: 6b17faa66a43a19fb8350cd19ab7ec538f9d068b
```