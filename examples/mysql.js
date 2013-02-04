user = "example"
database = "example"

mysql_database(
	database,									// database name
	{
		username: "root",
		password: "",
		host: "127.0.0.1",
		port: 3306
	}
)

mysql_user(
	user,													// Use the user variable set above
	"testing",										// Set the password (unencrypted)
	"127.0.0.1",									// Host that the user can connect from
	database,											// Database the user has access to
	{
		username: "root",						// default is root
		password: "",								// default is blank
		host: "127.0.0.1",					// default is localhost
		port: 3306									// default is 3306
	}
)

