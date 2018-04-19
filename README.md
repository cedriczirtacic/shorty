# shorty
Disposable short url server intended to be used from the console (kinda like _tinyurl_).
## Description
**shorty** has two interfaces:
 1) Listens for TCP connections and transforms the provided URL to a short url.
 2) Listens for HTTPS requests with the given URL (step 1) and redirects the user to the original URL.

Short URL gets destroyed after the expiration time is exceeded.
**shorty** is not using any third-party package so `go get` is not needed.
## Example:

Launch the server:
```bash
$ go build
$ ./shorty
[shorty] 2018/04/18 14:29:59.357699 setting up the shorty server on: localhost:6666
[shorty] 2018/04/18 14:29:59.357734 setting up httpd server on: https://localhost:443
```
Request a URL and test the redirection:
```bash
$ echo "http://google.com" | nc localhost 6666 
https://localhost/vKgGBO
$ curl -skL https://localhost/vKgGBO
<!doctype html><html itemscope="" itemtype="http://schema.org/WebPage" lang="es-419"><head><meta content="text/html; charset=UTF-8" http-equiv="Content-Type"><meta content="/images/branding/googleg/1x/googleg_standard_color_128dp.png" itemprop="image"><title>Google</title>...
...
```
On server side you will confirm the transaction:
```bash
[shorty] 2018/04/18 14:30:56.809310 processing connection from 127.0.0.1:49846
[shorty] 2018/04/18 14:30:56.809569 generated URL: https://localhost/vKgGBO
```
To launch the HTTPd server you'll need to generate or specify a key and certificate. For this, `build_cert.sh` is provided but you can choose your own way to do it.

In case the file __index.html__ is present, it will be loaded if / or /index.html is requested.

## Options
```bash
Usage of ./shorty:
  -L int
    	Length of url ID. (default 6)
  -h string
    	Shorty server address. (default "localhost")
  -l int
    	Short url lifespan. (default 10)
  -p int
    	Shorty server port. (default 6666)
  -wcert string
    	Shorty httpd ssl certificate path. (default "shorty.crt")
  -wd string
    	Shorty httpd server domain to use. (default "localhost")
  -wh string
    	Shorty httpd server address. (default "localhost")
  -wkey string
    	Shorty httpd ssl private key path. (default "shorty.key")
  -wp int
    	Shorty httpd server port. (default 443)
```

## Notes
 * Why HTTPS only? Because that should be a must.

