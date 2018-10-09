# auth-static

> Protect static files to be accessible to logged-in users only.

The files are served by a web server that supports the `X-Accel-Redirect` HTTP header.
Authentication is handled by a web app written in Go.

The following parts are included:
- command line tool to add new users which are stored in a SQLite database
- signup handler that allows new users to set their password
- signin handler that allows users to log in
- authentication handler that ensures the user is logged in and that tells the web server to serve the requested static file

## Prerequisites

- [Caddy](https://caddyserver.com/) (for alternatives see [Web server](#web-server))
- [Go](https://golang.org/)
- [dep](https://golang.github.io/dep/docs/installation.html)

## Install

Clone this repo, change into its dir, install its dependencies:

    $ dep ensure

Install the binaries. The output dir specified by `-o` has to be in your `$PATH`.

    $ go build -o ~/bin/as-initdb ./cmd/initdb/
    $ go build -o ~/bin/as-createuser ./cmd/createuser/
    $ go build -o ~/bin/as-web ./cmd/web/
    $ go build -o ~/bin/as-genkey ./cmd/genkey/

## Example

Change into the `example` directory.

Start the web server:

    $ caddy

Create the database, in another shell:

    $ as-initdb

This creates a `prod.db` in the current working directory.
To put it somewhere else or use a different name use the `-dsn` flag:

    $ as-initdb -dsn "/path/to/my.db"

See go-sqlite3's [SQLiteDriver.Open](https://godoc.org/github.com/mattn/go-sqlite3#SQLiteDriver.Open) for accepted values.

Generate key pair for cookie security:

    $ as-genkey
    8cb...
    3cf...

Persist both keys somewhere save.

Start the app:

    $ as-web -hashkey 8cb... -blockkey 3cf...

_Note: use the `-dsn` flag if the database file is not `prod.db` in the current working directory._

In production e.g. if using HTTPS also add the `Secure` flag to the cookie:

    $ as-web -hashkey 8cb... -blockkey 3cf... -secure

Create the first user, in another shell:

    $ as-createuser -email me@example.com
    successfully created user with email "me@example.com" and code "e80ef0a04db3597e09fee4e958ca12b1"

_Note: use the `-dsn` flag if the database file is not `prod.db` in the current working directory._

Use the code to create a URL:

    http://localhost:8080/signup/e80ef0a04db3597e09fee4e958ca12b1

That's the URL to initially set a password. After that the user will be redirected to the protected area.
Users can sign in on http://localhost:8080/signin.
http://localhost:8080/ is public.
Everything in the `internal` directory is protected and accessible only to authenticated requests via `private` URL path: http://localhost:8080/private/main.html.

## Web server

Any webserver that supports the `X-Accel-Redirect` or `X-Sendfile` HTTP headers can be used. For example:

- [http.internal](https://caddyserver.com/docs/internal) - Caddy, see `./example/Caddyfile` for example configuration.
- [X-Accel](https://www.nginx.com/resources/wiki/start/topics/examples/x-accel/) and [XSendfile](https://www.nginx.com/resources/wiki/start/topics/examples/xsendfile/) - nginx
- [mod_xsendfile](https://tn123.org/mod_xsendfile/) - Apache
