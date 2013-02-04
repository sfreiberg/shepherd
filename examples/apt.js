// install one package with no options
apt("example-pkg")

// install multiple packages with no options
apt(["example-pkg", "example-pkg2"])

// - or -
pkgs = ["example-pkg", "example-pkg2"]
apt(pkgs)
