user(
	"user1",
	{
		homeDir: "/home/user1",
		shell: "/bin/bash",
		uid: 500,
		gid: 500,
		password: "$6$somegiantcrazyhashthathasbeencryptedorsomething",
		// createDefaultGroup: false, // <- defaults to true
		groups: ["sudo"]
	}
)
