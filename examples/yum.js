// install one package with no options
yum("example-pkg")

// install multiple packages with no options
yum(["example-pkg", "example-pkg2"])

// - or -
pkgs = ["example-pkg", "example-pkg2"]
yum(pkgs)

// install package with options
yum("example-pkg", {groupinstall: true})
