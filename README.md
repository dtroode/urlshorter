# go-musthave-shortener-tpl

**развитие**
- [ ] оптимизировать мидлварь DeGzip, чтоб не аллоцировать новый ридер на каждый запрос. для этого можно использовать [sync.Pool](https://pkg.go.dev/sync#Pool), но нужно лучше разбираться в примитивах синхронизации. опираться на то, как сделано в библиотеке [gziphandler](https://github.com/NYTimes/gziphandler)
