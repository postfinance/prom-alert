# prom-alert

[![Go Report Card](https://goreportcard.com/badge/github.com/postfinance/prom-alert)](https://goreportcard.com/report/github.com/postfinance/prom-alert)
[![Coverage Status](https://coveralls.io/repos/github/postfinance/prom-alert/badge.svg?branch=master)](https://coveralls.io/github/postfinance/prom-alert?branch=master)
[![Build](https://github.com/postfinance/prom-alert/workflows/build/badge.svg)](https://github.com/postfinance/prom-alert/actions?query=workflow%3Abuild)

This tool lets you create prometheus test alerts.

## Usage

```
prom-alert -labels team=linux,severity=warning
```

You can hit `ctrl+c` to stop the alert firing.
