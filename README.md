Shepherd
========

Shepherd is configuration and server management wrapped in an easy to use, straightforward package.

Install
=======

At the moment there are no binaries available. They will be coming shortly.

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
	user,													// Use the user variable set above
	"password",										// Set the password (unencrypted)
	{
		username: "postgres",				// default is root
		password: "",								// default is blank
		host: "localhost",					// default is localhost
		port: 5432,									// default is 3306
		ssl: false									// do not use ssl
	}
)

pg_database(
	"example",									// database name
	user,												// database owner
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
