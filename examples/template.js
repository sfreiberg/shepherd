template(
	"./examples/shepherd.tpl",		// The template file
	"./examples/shepherd.conf",		// The destination
	{port: "80"},									// Context
	{}														// {perms: "0640", owner: "root", group: "root"}
)