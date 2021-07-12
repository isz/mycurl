# Утилита mycurl

Простой HTTP клиент по типу curl.

Поддерживается только HTTP

Реализованы методы GET, POST

Сборка:
```shell
go build -o mycurl cmd/mycurl/main.go
```

Использование:
```shell
./mycurl -v -m GET http://host:port/path
./mycurl -v -m POST -c application/json http://host:port/path '{"key":"value"}'
```

Краткая справка:
```shell
./mycurl --help
```

*Пользуйтесь на здоровье*
