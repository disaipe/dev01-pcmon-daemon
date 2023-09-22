# Dev01 PC status monitor daemon

```shell
dev01-pcmon-daemon.exe -serve -addr=[:8090] -app.url=http://dev01.com/api/pcmon -app.secret=[secret word]
```

```shell
# install as Windows service
dev01-pcmon-daemon.exe -srv.install

# start as Windows service
dev01-pcmon-daemon.exe -srv ...

# uninstall Windows service
dev01-pcmon-daemon.exe -srv.uninstall
```
