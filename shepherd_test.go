package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/robertkrimen/otto"
)

// withCwd points the global cwd (used to resolve files/ and templates/)
// at dir for the duration of the test.
func withCwd(t *testing.T, dir string) {
	t.Helper()
	old := cwd
	cwd = dir
	t.Cleanup(func() { cwd = old })
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestPgConnString(t *testing.T) {
	tests := []struct {
		ci   PgConnInfo
		want string
	}{
		{PgConnInfo{}, "dbname=postgres sslmode=disable"},
		{
			PgConnInfo{Username: "postgres", Password: "pw", Host: "db", Port: 5433, Ssl: true},
			"dbname=postgres user=postgres password=pw host=db port=5433 sslmode=require",
		},
	}
	for _, tt := range tests {
		if got := tt.ci.GetConnString(); got != tt.want {
			t.Errorf("GetConnString(%+v) = %q, want %q", tt.ci, got, tt.want)
		}
	}
}

func TestPgComputePassword(t *testing.T) {
	// PostgreSQL md5 passwords are "md5" + md5(password + username).
	u := NewPgUser("example", "password")
	if got, want := u.ComputePassword(), "md5be86a79bf2043622d58d5453c47d4860"; got != want {
		t.Errorf("ComputePassword() = %q, want %q", got, want)
	}
	pre := NewPgUser("example", "md5be86a79bf2043622d58d5453c47d4860")
	if got := pre.ComputePassword(); got != pre.Password {
		t.Errorf("pre-hashed password was re-hashed: %q", got)
	}
}

func TestCommand(t *testing.T) {
	c := NewCommand("echo hello")
	c.Run()
	if !c.Success || strings.TrimSpace(c.Output) != "hello" || c.Error != "" {
		t.Fatalf("result = %+v", c.Result)
	}

	c = NewCommand("echo oops >&2; exit 3")
	c.Run()
	if c.Success || c.Error == "" || strings.TrimSpace(c.Output) != "oops" {
		t.Fatalf("result = %+v", c.Result)
	}
}

func TestSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	other := filepath.Join(dir, "other")
	link := filepath.Join(dir, "link")
	writeFile(t, target, "x")
	writeFile(t, other, "y")

	s := NewSymlink(target, link)
	s.Run()
	if !s.Success || !s.Changed {
		t.Fatalf("create: %+v", s.Result)
	}
	if got, _ := os.Readlink(link); got != target {
		t.Fatalf("link points to %q", got)
	}

	s = NewSymlink(target, link)
	s.Run()
	if !s.Success || s.Changed {
		t.Fatalf("idempotent run: %+v", s.Result)
	}

	s = NewSymlink(other, link)
	s.Run()
	if !s.Success || !s.Changed {
		t.Fatalf("retarget: %+v", s.Result)
	}
	if got, _ := os.Readlink(link); got != other {
		t.Fatalf("link points to %q, want %q", got, other)
	}
}

func TestFile(t *testing.T) {
	dir := t.TempDir()
	withCwd(t, dir)
	writeFile(t, filepath.Join(dir, "files", "app.conf"), "setting=1\n")
	dest := filepath.Join(dir, "app.conf")

	f := NewFile("app.conf", dest)
	f.Run()
	if !f.Success || !f.Changed || readFile(t, dest) != "setting=1\n" {
		t.Fatalf("first run: %+v", f.Result)
	}

	f = NewFile("app.conf", dest)
	f.Run()
	if !f.Success || f.Changed {
		t.Fatalf("unchanged run: %+v", f.Result)
	}

	writeFile(t, dest, "tampered")
	f = NewFile("app.conf", dest)
	f.Run()
	if !f.Success || !f.Changed || readFile(t, dest) != "setting=1\n" {
		t.Fatalf("restore run: %+v", f.Result)
	}

	f = NewFile("missing.conf", filepath.Join(dir, "x"))
	f.Run()
	if f.Success || f.Error == "" {
		t.Fatalf("missing source: %+v", f.Result)
	}
}

func TestDirectory(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "a", "b")

	d := NewDirectory(dest)
	d.Run()
	if !d.Success || !d.Changed {
		t.Fatalf("create: %+v", d.Result)
	}
	if fi, err := os.Stat(dest); err != nil || !fi.IsDir() {
		t.Fatalf("directory not created: %v", err)
	}

	d = NewDirectory(dest)
	d.Run()
	if !d.Success || d.Changed {
		t.Fatalf("existing: %+v", d.Result)
	}
}

func TestDirectoryReplacesFile(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "d")
	writeFile(t, dest, "not a dir")

	d := NewDirectory(dest)
	d.Run()
	if !d.Success || !d.Changed {
		t.Fatalf("result = %+v", d.Result)
	}
	if fi, err := os.Stat(dest); err != nil || !fi.IsDir() {
		t.Fatalf("file was not replaced by a directory: %v", err)
	}
}

func TestTemplate(t *testing.T) {
	dir := t.TempDir()
	withCwd(t, dir)
	writeFile(t, filepath.Join(dir, "templates", "app.tpl"),
		"[server]\nport={{ port }}\nbanner=\"{{ banner }}\"\n")
	dest := filepath.Join(dir, "app.conf")
	ctx := map[string]any{"port": "80", "banner": `<b>Tom & "Jerry"</b>`}

	tpl := NewTemplate("app.tpl", dest)
	tpl.Context = ctx
	tpl.Run()
	if !tpl.Success || !tpl.Changed {
		t.Fatalf("first run: %+v", tpl.Result)
	}
	// Config files must not be HTML-escaped.
	want := "[server]\nport=80\nbanner=\"<b>Tom & \"Jerry\"</b>\"\n"
	if got := readFile(t, dest); got != want {
		t.Fatalf("rendered %q, want %q", got, want)
	}

	tpl = NewTemplate("app.tpl", dest)
	tpl.Context = ctx
	tpl.Run()
	if !tpl.Success || tpl.Changed {
		t.Fatalf("unchanged run: %+v", tpl.Result)
	}
}

func TestAptHelpers(t *testing.T) {
	a := NewApt()
	a.AddPackages([]string{"nginx", "curl"})
	a.Packages["curl"].Installed = true
	a.StatusUpdated = true

	missing, err := a.MissingPkgNames()
	if err != nil || len(missing) != 1 || missing[0] != "nginx" {
		t.Fatalf("MissingPkgNames() = %v, %v", missing, err)
	}
	if ok, _ := a.MissingPkgs(); !ok {
		t.Fatal("MissingPkgs() = false, want true")
	}
	if a.NoInstallRecommendsOpt() != "" {
		t.Fatal("NoInstallRecommendsOpt should be empty by default")
	}
	a.NoInstallRecommends = true
	if a.NoInstallRecommendsOpt() != "--no-install-recommends" {
		t.Fatal("NoInstallRecommendsOpt not set")
	}
}

func TestUserGroups(t *testing.T) {
	u := NewUser("deploy")
	if !u.CreateDefaultGroup {
		t.Fatal("CreateDefaultGroup should default to true")
	}
	u.AddSecondaryGroup("sudo")
	u.AddSecondaryGroup("docker")
	if got := u.SecondaryGroupsAsString(); got != "sudo,docker" {
		t.Fatalf("SecondaryGroupsAsString() = %q", got)
	}
}

func TestConvertJSValues(t *testing.T) {
	vm := otto.New()

	v, err := vm.Run(`({username: "pg", password: "pw", host: "db", port: 5433, ssl: true})`)
	if err != nil {
		t.Fatal(err)
	}
	ci := convertValToPgConnInfo(&v)
	if ci.Username != "pg" || ci.Password != "pw" || ci.Host != "db" || ci.Port != 5433 || !ci.Ssl {
		t.Fatalf("PgConnInfo = %+v", ci)
	}
	mi := convertValToMysqlConnInfo(&v)
	if mi.Username != "pg" || mi.Port != 5433 {
		t.Fatalf("MysqlConnInfo = %+v", mi)
	}

	v, err = vm.Run(`({perms: "0640", user: "root"})`)
	if err != nil {
		t.Fatal(err)
	}
	m := convertValToMap(&v)
	if m["perms"] != "0640" || m["user"] != "root" {
		t.Fatalf("map = %v", m)
	}

	undef := otto.UndefinedValue()
	if m := convertValToMap(&undef); m == nil || len(m) != 0 {
		t.Fatalf("undefined should convert to an empty map, got %v", m)
	}
}

func TestJavascriptBindings(t *testing.T) {
	initJavascript()
	dir := t.TempDir()
	vm := js
	vm.Set("target", filepath.Join(dir, "made", "by", "js"))

	v, err := vm.Run(`directory(target).success`)
	if err != nil {
		t.Fatal(err)
	}
	if ok, _ := v.ToBoolean(); !ok {
		t.Fatal("directory() did not report success")
	}
	if _, err := os.Stat(filepath.Join(dir, "made", "by", "js")); err != nil {
		t.Fatal(err)
	}

	v, err = vm.Run(`command("echo hi").output`)
	if err != nil {
		t.Fatal(err)
	}
	if s, _ := v.ToString(); strings.TrimSpace(s) != "hi" {
		t.Fatalf("command() output = %q", s)
	}

	v, err = vm.Run(`typeof facts.hostname`)
	if err != nil {
		t.Fatal(err)
	}
	if s, _ := v.ToString(); s != "string" {
		t.Fatalf("facts.hostname has type %q", s)
	}
}

func TestFilePermsIdempotent(t *testing.T) {
	// "0640" must match stat's "640", otherwise every run chmods the file
	// and reports a change.
	dir := t.TempDir()
	withCwd(t, dir)
	writeFile(t, filepath.Join(dir, "files", "secret.conf"), "x")
	dest := filepath.Join(dir, "secret.conf")

	for i, wantChanged := range []bool{true, false} {
		f := NewFile("secret.conf", dest)
		f.Options = map[string]any{"perms": "0640"}
		f.Run()
		if !f.Success || f.Changed != wantChanged {
			t.Fatalf("run %d: %+v", i, f.Result)
		}
	}
	fi, err := os.Stat(dest)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0o640 {
		t.Fatalf("mode = %v, want 0640", fi.Mode().Perm())
	}
}

func TestSamePerms(t *testing.T) {
	for _, tt := range []struct {
		a, b string
		want bool
	}{
		{"640", "0640", true},
		{"755", "755", true},
		{"0", "000", true},
		{"644", "0640", false},
		{"640", "u+r", false},
	} {
		if got := samePerms(tt.a, tt.b); got != tt.want {
			t.Errorf("samePerms(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestPackageManagerArgs(t *testing.T) {
	a := NewApt()
	got := strings.Join(a.aptArgs([]string{"nginx", "foo; rm -rf /"}), "|")
	if want := "-y|-q|install|nginx|foo; rm -rf /"; got != want {
		t.Errorf("aptArgs = %q, want %q", got, want)
	}
	a.NoInstallRecommends = true
	got = strings.Join(a.aptArgs([]string{"nginx"}), "|")
	if want := "-y|-q|--no-install-recommends|install|nginx"; got != want {
		t.Errorf("aptArgs = %q, want %q", got, want)
	}

	y := NewYum()
	y.AddPackage("httpd")
	y.AddPackage("mod_ssl")
	y.Packages["mod_ssl"].Installed = true
	if got := strings.Join(y.yumArgs(), "|"); got != "install|-y|-q|httpd" {
		t.Errorf("yumArgs = %q", got)
	}
	y = NewYum()
	y.Groupinstall = true
	y.AddPackage("Development Tools")
	if got := y.yumArgs(); len(got) != 4 || got[0] != "groupinstall" || got[3] != "Development Tools" {
		t.Errorf("yumArgs = %q", got)
	}
}

func TestMysqlIdent(t *testing.T) {
	for in, want := range map[string]string{
		"app":            "`app`",
		"my-db":          "`my-db`",
		"x`; DROP x; --": "`x``; DROP x; --`",
	} {
		if got := mysqlIdent(in); got != want {
			t.Errorf("mysqlIdent(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestConvertValToMapNonObject(t *testing.T) {
	vm := otto.New()
	v, err := vm.Run(`["a", "b"]`)
	if err != nil {
		t.Fatal(err)
	}
	// An array is an object in JS but not a map; this used to panic.
	if m := convertValToMap(&v); m == nil || len(m) != 0 {
		t.Fatalf("got %v, want empty map", m)
	}
}
