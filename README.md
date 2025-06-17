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
