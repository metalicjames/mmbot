# mmbot

mmbot is a simple market maker bot that implements a constant interval strategy to provide a spread. It is currently
implemented to work on [Vertpig](https://vertpig.com) but will probably work on Bittrex with very little modification.
The exchange interface itself is generic so any exchange can be supported in theory if an API interface is written for it.
The interface is very straightforward and described in the `exchange.go` file.

## Getting started

mmbot is written in Go. I use Go 1.9 but it will almost certainly work with earlier and later versions. Assuming you 
have Go installed:

```
go get github.com/metalicjames/mmbot
```

Now retrieve an API key from the exchange you want to use (currently only Vertpig at this time) and fill in `config.json`
with the key and secret. The default config file contains sane defaults for each of the markets.

## No warranty

This software is under the GPL and as such has no warranty. This means I am not responsible for anything you do with
this software, including potentially losing money. Please read the license to ensure you understand it before using
the software.
