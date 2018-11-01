# HTTP status check

Concurrently checks the HTTP status code of all links in a CSV or a sitemap XML file and output the results (status code, request time and potential redirect url) in another CSV file

You may also change the scheme/host of the links from CLI (ie: transform http://exemple.com/my/url to https://staging.exemple.com/my/url)


## Usage

```
$ http-status-check --help

Usage: http-status-check [-c=<concurrency>] [-i=<input-file-path>] [-o=<output-file-path>] [-n=<new-uri>]

CLI tool to concurrently checks URLs from a CSV or a sitemap XML file and check HTTP status code

Options:
  -c, --concurrency=5         Concurrency
  -i, --input="input.csv"     Input CSV or sitemap XML file path
  -o, --output="output.csv"   Output CSV file path
  -n, --newuri=""             New URI for scheme/host replacements (ie: https://staging.exemple.com/)
```


## Exemples

### Simple exemple

```
http-status-check -i exemple.csv
http-status-check -i sitemap.xml
```

### Complete exemple

```
http-status-check -i exemple.csv -o raw-github-test.csv -c 2 -n https://raw.github.com/
http-status-check -i sitemap.xml -o raw-github-test.csv -c 2 -n https://raw.github.com/
```


## Installation from binaries

See the [GitHub releases](https://github.com/Benoth/http-status-check/releases)


## Installation from sources

### Install Go

```
bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
source ~/.gvm/scripts/gvm
sudo apt-get install bison
gvm install go1.4 -B
gvm use go1.4
export GOROOT_BOOTSTRAP=$GOROOT
gvm install go1.7 --prefer-binary
gvm use go1.7 --default
```

### Install packages

```
go get github.com/jawher/mow.cli
go get github.com/parnurzeal/gorequest
go get gopkg.in/go-playground/pool.v3
```

### Cross compile

Use gox :

```
go get github.com/mitchellh/gox
gox
```
