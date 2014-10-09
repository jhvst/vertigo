vertigo [![wercker status](https://app.wercker.com/status/e1f07b85320f902313d32fec503c5017/s/master "wercker status")](https://app.wercker.com/project/bykey/e1f07b85320f902313d32fec503c5017)
=======

Vertigo aims to be portable and easy to customize publishing platform.

Once more stable, Vertigo is supposed to be available in simple executable file for platforms supported by Go. Any unnecessary 3rd party packages are subject to be removed to further increase easy portability.

Vertigo is developed with frontend JavaScript frameworks in mind, although Vertigo doesn't itself ship with one. All actions which can be done on the site can also done with the JSON API. Even though Vertigo is programmed API-first, it does not itself rely on AJAX calls, but makes the corresponding calls server side. This makes it possible to ship Vertigo nearly JavaScript-free (the only page using JavaScript is the text editor), letting the developer decide what frontend tools to use. Vertigo also ships without any CSS frameworks, so you can start theming from ground-zero. Templating engine is `html/template` of Go's standard library, which looks much same as Mustache. You don't need to know Go to edit how Vertigo looks.

On the backend the server makes a heavy use of [Martini](http://martini.codegangsta.io/), which makes programming web services in Go bit easier. The database driver is [gorm](https://github.com/jinzhu/gorm), which means Vertigo supports SQLite, MySQL and PostgreSQL out of the box. By default SQLite is used.

##Demo

See [my personal website](http://www.juusohaavisto.com/)

##Install instructions

1. Install Go (I recommend using [gvm](https://github.com/moovweb/gvm))
2. `go get github.com/tools/godep`
3. `git clone github.com/9uuso/vertigo`
4. `cd vertigo && go get ./ && godep go build`
5. Start Vertigo `PORT="80" MARTINI_ENV="production" ./vertigo`

##Screenshots

![](http://i.imgur.com/EGlBhjP.png)
![](http://i.imgur.com/0AfvQnW.png)
![](http://i.imgur.com/AeC9xml.png)
![](http://i.imgur.com/rDlM9IX.png)
![](http://i.imgur.com/EwFcRfq.png)

##Features

- Create and read users
- Create, read, update and delete posts
- Session control
- Password protected user page
- Basic homepage which lists all published posts from all users
- JSON API
- Produce HTML5 compliant code with text editor (divitism, but I've decided to not battle against contenteditable)
- Failover auto-saving to LocalStorage
- Searching of posts
- RSS and Atom feeds
- Password recovery with (forced :new_moon_with_face:) Mailgun integration
- Installation wizard

##License

MIT
