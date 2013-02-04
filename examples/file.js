result = file(
	"./apache2.conf.large",		// The source file
	"./apache.conf",					// The destination
	{perms: "640", owner: "www-data", group: "www-data"}
)
