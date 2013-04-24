package main

import (
	"github.com/robertkrimen/otto"

	"fmt"
)

func convertResultToObject(r *Result) otto.Value {
	result, err := js.Object(`result = {}`)
	if err != nil {
		return otto.UndefinedValue()
	}

	result.Set("success", r.Success)
	result.Set("output", r.Output)
	result.Set("changed", r.Changed)
	result.Set("error", r.Error)

	return result.Value()
}

func displayError(actionName string, r *Result) {
	if r.Success == false {
		fmt.Printf("\x1b[31;1m*************** ERROR: %v ***************\x1b[0m\n", actionName)
		fmt.Printf("\x1b[31;1m%v\x1b[0m\n", r.Error)
		fmt.Printf("\x1b[31;1m%v\x1b[0m\n", r.Output)
	}
}

func directory(call otto.FunctionCall) otto.Value {
	path, _ := call.Argument(0).ToString()
	options := call.Argument(1)
	directory := NewDirectory(path)

	directory.Options = convertValToMap(&options)

	directory.Run()
	displayError("creating directory", &directory.Result)
	return convertResultToObject(&directory.Result)
}

func symlink(call otto.FunctionCall) otto.Value {
	source, _ := call.Argument(0).ToString()
	destination, _ := call.Argument(1).ToString()
	symlink := NewSymlink(source, destination)
	symlink.Run()
	displayError("creating symlink", &symlink.Result)
	return convertResultToObject(&symlink.Result)
}

func upstart(call otto.FunctionCall) otto.Value {
	svc, _ := call.Argument(0).ToString()
	action, _ := call.Argument(1).ToString()
	upstart := NewUpstart(svc, action)

	upstart.Run()
	displayError("with upstart", &upstart.Result)
	return convertResultToObject(&upstart.Result)
}

func template(call otto.FunctionCall) otto.Value {
	source, _ := call.Argument(0).ToString()
	destination, _ := call.Argument(1).ToString()
	context := call.Argument(2)
	options := call.Argument(3)

	t := NewTemplate(source, destination)
	t.Context = convertValToMap(&context)
	t.Options = convertValToMap(&options)

	t.Run()
	displayError("creating template", &t.Result)
	return convertResultToObject(&t.Result)
}

func apt(call otto.FunctionCall) otto.Value {
	// Need at least one argument
	if call.Argument(0).IsUndefined() {
		e := "No arguments given to apt() method"
		r := &Result{Success: false, Changed: false, Error: e}
		return convertResultToObject(r)
	}

	p := NewApt()
	// Add a single package
	pkgs := call.Argument(0)
	if pkgs.IsString() {
		pkgName, err := pkgs.ToString()
		if err != nil {
			r := &Result{Success: false, Changed: false, Error: err.Error()}
			return convertResultToObject(r)
		}
		p.AddPackage(pkgName)
		// Add multiple packages
	} else if pkgs.Class() == "Array" {
		pkgNames, err := pkgs.Export()
		if err != nil {
			r := &Result{Success: false, Changed: false, Error: err.Error()}
			return convertResultToObject(r)
		}
		for _, pkgName := range pkgNames.([]interface{}) {
			p.AddPackage(pkgName.(string))
		}
	}

	// Check for optional arguments
	options := call.Argument(1)
	if options.IsDefined() && options.IsObject() {
		noInstallRecommendsVal, err := options.Object().Get("noInstallRecommends")
		if err == nil && noInstallRecommendsVal.IsBoolean() {
			noInstallRecommends, _ := noInstallRecommendsVal.ToBoolean()
			if noInstallRecommends {
				p.NoInstallRecommends = true
			}
		}
	}

	p.Run()
	displayError("installing packages", &p.Result)
	return convertResultToObject(&p.Result)
}

func yum(call otto.FunctionCall) otto.Value {
	// Need at least one argument
	if call.Argument(0).IsUndefined() {
		e := "No arguments given to yum() method"
		r := &Result{Success: false, Changed: false, Error: e}
		return convertResultToObject(r)
	}

	p := NewYum()
	// Add a single package
	pkgs := call.Argument(0)
	if pkgs.IsString() {
		pkgName, err := pkgs.ToString()
		if err != nil {
			r := &Result{Success: false, Changed: false, Error: err.Error()}
			return convertResultToObject(r)
		}
		p.AddPackage(pkgName)
		// Add multiple packages
	} else if pkgs.Class() == "Array" {
		pkgNames, err := pkgs.Export()
		if err != nil {
			r := &Result{Success: false, Changed: false, Error: err.Error()}
			return convertResultToObject(r)
		}
		for _, pkgName := range pkgNames.([]interface{}) {
			p.AddPackage(pkgName.(string))
		}
	}

	// Check for groupinstall
	options := call.Argument(1)
	if options.IsDefined() && options.IsObject() {
		groupinstallValue, err := options.Object().Get("groupinstall")
		if err == nil && groupinstallValue.IsBoolean() {
			groupinstall, _ := groupinstallValue.ToBoolean()
			if groupinstall {
				p.Groupinstall = true
			}
		}
	}

	p.Run()
	displayError("installing packages", &p.Result)
	return convertResultToObject(&p.Result)
}

func command(call otto.FunctionCall) otto.Value {
	cmd, _ := call.Argument(0).ToString()
	c := NewCommand(cmd)
	c.Run()
	displayError("executing command", &c.Result)
	return convertResultToObject(&c.Result)
}

func user(call otto.FunctionCall) otto.Value {
	username, _ := call.Argument(0).ToString()
	u := NewUser(username)

	options := call.Argument(1)
	if options.IsDefined() && options.IsObject() {
		obj := options.Object()

		// get homeDir
		if h, err := obj.Get("homeDir"); err == nil && h.IsString() {
			u.HomeDir, _ = h.ToString()
		}

		// get shell
		if s, err := obj.Get("shell"); err == nil && s.IsString() {
			u.Shell, _ = s.ToString()
		}

		// get uid
		if uid, err := obj.Get("uid"); err == nil && uid.IsNumber() {
			u.UserId, _ = uid.ToInteger()
		}

		// get gid
		if g, err := obj.Get("gid"); err == nil && g.IsNumber() {
			u.GroupId, _ = g.ToInteger()
		}

		// get password
		if p, err := obj.Get("password"); err == nil && p.IsString() {
			u.Password, _ = p.ToString()
		}

		// get secondary groups
		secondaryGroups, err := obj.Get("groups")
		if err == nil && secondaryGroups.IsString() {
			secondaryGroup, err := secondaryGroups.ToString()
			if err != nil {
				r := &Result{Success: false, Changed: false, Error: err.Error()}
				return convertResultToObject(r)
			}
			u.AddSecondaryGroup(secondaryGroup)
			// Add multiple secondary groups
		} else if err == nil && secondaryGroups.Class() == "Array" {
			secondaryGroupNames, err := secondaryGroups.Export()
			if err != nil {
				r := &Result{Success: false, Changed: false, Error: err.Error()}
				return convertResultToObject(r)
			}
			for _, groupName := range secondaryGroupNames.([]interface{}) {
				u.AddSecondaryGroup(groupName.(string))
			}
		}
	}

	u.Run()
	displayError("creating user", &u.Result)
	return convertResultToObject(&u.Result)
}

func pgUser(call otto.FunctionCall) otto.Value {
	username, _ := call.Argument(0).ToString()
	password, _ := call.Argument(1).ToString()
	user := NewPgUser(username, password)

	options := call.Argument(2)
	user.PgConnInfo = convertValToPgConnInfo(&options)

	user.Run()
	displayError("creating postgres user", &user.Result)
	return convertResultToObject(&user.Result)
}

func pgDatabase(call otto.FunctionCall) otto.Value {
	name, _ := call.Argument(0).ToString()
	owner, _ := call.Argument(1).ToString()
	db := NewPgDatabase(name, owner)

	options := call.Argument(2)
	db.PgConnInfo = convertValToPgConnInfo(&options)

	db.Run()
	displayError("creating postgres database", &db.Result)
	return convertResultToObject(&db.Result)
}

func mysqlUser(call otto.FunctionCall) otto.Value {
	username, _ := call.Argument(0).ToString()
	password, _ := call.Argument(1).ToString()
	hostname, _ := call.Argument(2).ToString()
	database, _ := call.Argument(3).ToString()
	user := NewMysqlUser(username, password, hostname, database)

	options := call.Argument(4)
	user.MysqlConnInfo = convertValToMysqlConnInfo(&options)

	user.Run()
	displayError("creating mysql user", &user.Result)
	return convertResultToObject(&user.Result)
}

func mysqlDatabase(call otto.FunctionCall) otto.Value {
	name, _ := call.Argument(0).ToString()
	db := NewMysqlDatabase(name)

	options := call.Argument(1)
	db.MysqlConnInfo = convertValToMysqlConnInfo(&options)

	db.Run()
	displayError("creating mysql database", &db.Result)
	return convertResultToObject(&db.Result)
}

func sleep(call otto.FunctionCall) otto.Value {
	millis, _ := call.Argument(0).ToInteger()
	Sleep(millis)
	return otto.NullValue()
}

func file(call otto.FunctionCall) otto.Value {
	source, _ := call.Argument(0).ToString()
	destination, _ := call.Argument(1).ToString()
	options := call.Argument(2)

	f := NewFile(source, destination)
	f.Options = convertValToMap(&options)

	f.Run()
	displayError("copying file", &f.Result)
	return convertResultToObject(&f.Result)
}

func include(call otto.FunctionCall) otto.Value {
	module, _ := call.Argument(0).ToString()
	Include(module)
	return otto.NullValue()
}

func convertValToMap(v *otto.Value) map[string]interface{} {
	if v.IsDefined() && v.IsObject() {
		options, _ := v.Export()
		return options.(map[string]interface{})
	}
	return make(map[string]interface{})
}

func convertValToPgConnInfo(v *otto.Value) PgConnInfo {
	connInfo := PgConnInfo{}

	if v.IsDefined() && v.IsObject() {
		obj := v.Object()
		// get username
		if u, err := obj.Get("username"); err == nil && u.IsString() {
			connInfo.Username, _ = u.ToString()
		}

		// get password
		if p, err := obj.Get("password"); err == nil && p.IsString() {
			connInfo.Password, _ = p.ToString()
		}

		// get host
		if h, err := obj.Get("host"); err == nil && h.IsString() {
			connInfo.Host, _ = h.ToString()
		}

		// get port
		if p, err := obj.Get("port"); err == nil && p.IsNumber() {
			connInfo.Port, _ = p.ToInteger()
		}

		// get ssl
		if s, err := obj.Get("ssl"); err == nil && s.IsBoolean() {
			connInfo.Ssl, _ = s.ToBoolean()
		}
	}

	return connInfo
}

func convertValToMysqlConnInfo(v *otto.Value) *MysqlConnInfo {
	connInfo := &MysqlConnInfo{}

	if v.IsDefined() && v.IsObject() {
		obj := v.Object()
		// get username
		if u, err := obj.Get("username"); err == nil && u.IsString() {
			connInfo.Username, _ = u.ToString()
		}

		// get password
		if p, err := obj.Get("password"); err == nil && p.IsString() {
			connInfo.Password, _ = p.ToString()
		}

		// get host
		if h, err := obj.Get("host"); err == nil && h.IsString() {
			connInfo.Host, _ = h.ToString()
		}

		// get port
		if p, err := obj.Get("port"); err == nil && p.IsNumber() {
			connInfo.Port, _ = p.ToInteger()
		}
	}

	return connInfo
}
