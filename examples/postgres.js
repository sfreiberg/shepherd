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