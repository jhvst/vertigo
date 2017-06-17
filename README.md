vertigo
=======
[![Codeship Status for 9uuso/vertigo](https://img.shields.io/codeship/b2de9690-b16b-0132-08f1-3edef27c5b65/master.svg)](https://codeship.com/projects/69843) [![Deploy](https://img.shields.io/badge/heroku-deploy-green.svg)](https://heroku.com/deploy) [![GoDoc](https://godoc.org/github.com/9uuso/vertigo?status.svg)](https://godoc.org/github.com/9uuso/vertigo)

![Vertigo](http://i.imgur.com/ZnAQR6I.gif)

Vertigo is blogging platform similar to [Ghost](https://ghost.org), [Medium](https://medium.com) and [Tumblr](https://www.tumblr.com). Vertigo is written in Go and has fully featured JSON API and it can be run using single binary on all major operating systems like Windows, Linux and OSX.

The frontend code is powered by Go's `template/html` package, which is similar to Mustache.js. The template files are in plain HTML and JavaScript (vanilla) only appears on few pages. JavaScript is stripped down as much as possible to provide a better user experience on different devices. Vertigo also ships without any CSS frameworks, so it is easy to start customizing the frontend with the tools of your choice.

Thanks to the JSON API, it is easy to add your preferred JavaScript MVC on top of Vertigo. This means that you can create users, submit posts and read data without writing a single line of Go code. For example, one could write a single page application on top of Vertigo just by using JavaScript. Whether you want to take that path or just edit the HTML template files found in `/templates/` is up to you.

## Features

- Installation wizard
- JSON API
- SQLite and PostgreSQL support
- Fuzzy search
- Multiple account support
- Auto-saving of posts to LocalStorage
- RSS feeds
- Password recovery
- Markdown support

## Demo

See [my personal website](http://www.juusohaavisto.com/)

## Installation

Note: By default the HTTP server starts on port 3000. This can changed by declaring `PORT` environment variable or by passing one with the binary execution command.

### Downloading binaries

See [GitHub releases](https://github.com/9uuso/vertigo/releases).

### Heroku

[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy)

For advanced usage, see [Advanced Heroku deployment](https://github.com/9uuso/vertigo/wiki/Advanced-Heroku-deployment)

### Source

1. Install Go (I recommend using [gvm](https://github.com/moovweb/gvm))
2. `git clone https://github.com/9uuso/vertigo`
3. `cd vertigo && go build`
4. `PORT="80" ./vertigo`

### Docker
1. [Install docker](https://docs.docker.com/installation/)
2. `git clone https://github.com/9uuso/vertigo`
3. `cd vertigo`
4. `docker build -t "vertigo" .`
5. `docker run -d -p 80:80 vertigo`

### Environment variables
* `PORT` - the HTTP server port
* `SMTP_LOGIN` - address from which you want to send mail from. Example: postmaster@example.com
* `SMTP_PASSWORD` - Password for the mailbox defined with SMTP_LOGIN
* `SMTP_PORT` - SMTP port which to use to send email. Defaults to 587.
* `SMTP_SERVER` - SMTP server hostname or IP address. Example: smtp.example.org
* `DATABASE_URL` - database connection URL for PostgreSQL - if empty, SQLite will be used

## Contribute

Contributions are welcome, but before creating a pull request, please run your code trough `go fmt` and [`golint`](https://github.com/golang/lint). If the changes introduce new features, please  add tests for them. Try to squash your commits into one big one instead many small, to avoid unnecessary CI runs.

## Support

If you have any questions in mind, please file an issue.

## License

MIT
