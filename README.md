Shepherd
========

Shepherd is configuration and server management wrapped in an easy to use, straightforward package.

Install
=======

#### RPM based systems
```
# rpm -i http://netserious.com/downloads/shepherd-0.5.0-1.x86_64.rpm
```

#### Debian
```
# wget http://netserious.com/downloads/shepherd_0.5.0_amd64.deb
# dpkg -i shepherd_0.5.0_amd64.deb
```

#### Mac OS X

Download the latest pkg from http://netserious.com/downloads and double click on the installer.


Configuration Examples
======================

To run these examples save them in a file and run shepherd followed by the name of the file(s) to run. For example if you created a file named web_servers.js you would call:

```
shepherd web_servers.js
```

Install packages on Debian based systems:

```javascript
apt("example-pkg")
// Install multiple packages
apt(["example-pkg", "example-pkg2"])
```

Create 100 directories:

```javascript
for (var i = 0; i<100; i++) {
	directory("test"+i)
}
```

Create a PostgreSQL user and database:

```javascript
user = "example"

pg_user(
	user,
	"password",
	{
		username: "postgres",
		password: "",
		host: "localhost",
		port: 5432,
		ssl: false
	}
)

pg_database(
	"example",
	user,
	{
		username: "postgres",
		password: "",
		host: "localhost",
		port: 5432,
		ssl: false
	}
)
```

Create an operating system user:

```javascript
user(
	"user1",
	{
		homeDir: "/home/user1",
		shell: "/bin/bash",
		uid: 500,
		gid: 500
		password: "$6$somegiantcrazyhashthathasbeencryptedorsomething"
	}
)
```

Check out more examples in the examples directory.

Current Features
================

* Create your server configuration in javascript
* Create OS users
* Create MySQL databases and users
* Create PostgreSQL databases and users
* Enforce standard configuration files with templating
* Single native binary with low overhead

Planned Features
================

* Server implementation that can manage thousands of clients over http/https with very low overhead.
* Run ad-hoc tasks over one or thousands of servers instantly.
* *BSD support

Current Limitations
===================

* A limited set of actions are currently available.
* Only actively developed on Linux.
* No tests and limited testing at the moment.
* Little code documentation.
