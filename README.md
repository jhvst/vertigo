vertigo
=======

Currently under heavy development!

Vertigo aims to bring simpler way to publish simple websites or blogs, which are normally done with heavy content management systems like Wordpress. Basic ideaology is that functionality could be added with JavaScript even without touching the backend source code. Theme development should be distinct from backend code so that knowledge in Go would not be needed when developing themes.

Once more stable, Vertigo is supposed to be available in simple executable file for platforms supported by Go. Any unnecessary 3rd party packages are subject to be removed to further increase easy portability.

Vertigo is developed with frontend JavaScript frameworks in mind, although Vertigo doesn't itself ship with one. All actions which can be done on the site can also done with the JSON API. Even though Vertigo is programmed API-first, it does not itself rely on AJAX calls, but makes the corresponding calls server side. This makes it possible to ship Vertigo nearly JavaScript-free (the only page using JavaScript is the text editor), letting the developer decide what frontend tools to use.

On the backend the server makes a heavy use of [Martini](http://martini.codegangsta.io/), which makes programming web services in Go bit easier. Database of choice is [RethinkDB](http://rethinkdb.com/), which is forked version of MySQL. It serves data in JSON, but the query language comes with joins and plenty of other functionality not usually met with NoSQL databases. Unfortunately, there are very few SaaS products which provide RethinkDB, so you will most likely need to run one yourself. RethinkDB is the only part of the program which cannot be shipped in the executable.

##[Install instructions](https://github.com/9uuso/vertigo/releases/tag/v0.1-beta)

##Screenshots

![](http://i.imgur.com/EGlBhjP.png)
Homepage

![](http://i.imgur.com/0AfvQnW.png)
User home

![](http://i.imgur.com/AeC9xml.png)
Writing mode (nothing cropped out!)

![](http://i.imgur.com/rDlM9IX.png)
Post view mode (HTML content preserves it's form if edited)

![](http://i.imgur.com/EwFcRfq.png)
Content requested trough API

##What works

Since there is so much to do, I'd rather list some things that are currently working:

- Create and read users
- Create, read, update and delete posts
- Session control
- Password protected user page with CRUD options for each post
- Basic homepage which lists all posts from all users
- JSON API
- Produce HTML5 compliant code with text editor
- Auto-saving of posts to LocalStorage
- Search

##License

MIT