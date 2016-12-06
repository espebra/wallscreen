# wallscreen
Quick and dirty hack to get an icingaweb2 wallscreen

## Installation

```
go get github.com/espebra/wallscreen
make deps
make build
```

``$GOPATH/bin`` has to be in path for ``make build`` to succeed.

## Running

Standard way of running it:

```
bin/wallscreen-darwin-amd64  --baseurl https://monitor.example.com/icinga 
```

It is possible to add some custom request header with a custom value if needed, for example for authentication to the icinga API:

```
bin/wallscreen-darwin-amd64  --baseurl https://monitor.example.com/icinga --custom-header-name x-auth-user --custom-header-value letmein
```
