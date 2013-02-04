package actions

import (
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"

	"errors"
	"fmt"
	"regexp"
)

type MysqlConnInfo struct {
	Username string
	Password string
	Host     string
	Port     int64
}

type MysqlUser struct {
	Username string
	Password string
	Hostname string
	Database string
	Result
	MysqlConnInfo *MysqlConnInfo
}

func NewMysqlUser(user, pass, host, db string) *MysqlUser {
	connInfo := &MysqlConnInfo{"root", "", "127.0.0.1", 3306}
	mysqluser := &MysqlUser{
		Username:      user,
		Password:      pass,
		Hostname:      host,
		Database:      db,
		MysqlConnInfo: connInfo,
	}
	return mysqluser
}

func (mu *MysqlUser) Run() {
	db := mysql.New(
		"tcp",
		"",
		fmt.Sprintf("%v:%v", mu.MysqlConnInfo.Host, mu.MysqlConnInfo.Port),
		mu.MysqlConnInfo.Username,
		mu.MysqlConnInfo.Password,
		"mysql",
	)

	if err := db.Connect(); err != nil {
		mu.Result.Error = err.Error()
		return
	}

	userExists, err := mu.UserExists(db)
	if err != nil {
		mu.Result.Error = err.Error()
		return
	}

	if userExists != true {
		_, _, err := db.Query("CREATE USER '%v'@'%v' IDENTIFIED BY '%v'", mu.Username, mu.Hostname, mu.Password)
		if err != nil {
			mu.Result.Error = err.Error()
			return
		}
		mu.Result.Changed = true
	}

	grantExists, err := mu.GrantExists(db)
	if err != nil {
		mu.Result.Error = err.Error()
		return
	}
	if grantExists != true {
		_, _, err = db.Query("GRANT ALL ON %v.* TO '%v'@'%v'", mu.Database, mu.Username, mu.Hostname)
		if err != nil {
			mu.Result.Error = err.Error()
			return
		}
		mu.Result.Changed = true
	}

	mu.Result.Success = true
	return
}

func (mu *MysqlUser) UserExists(db mysql.Conn) (bool, error) {
	rows, _, err := db.Query(
		"SELECT COUNT(*) FROM mysql.user WHERE User = '%v' and Host = '%v' and Password = password('%v')",
		mu.Username,
		mu.Hostname,
		mu.Password,
	)
	if err != nil {
		return false, err
	}

	if len(rows) != 1 {
		return false, errors.New("Unexpected number of rows in UserExists()")
	}

	if rows[0].Int(0) > 0 {
		return true, nil
	}

	return false, nil
}

func (mu *MysqlUser) GrantExists(db mysql.Conn) (bool, error) {
	rows, _, err := db.Query(
		"SHOW GRANTS FOR '%v'@'%v'",
		mu.Username,
		mu.Hostname,
	)
	if err != nil {
		return false, err
	}

	if len(rows) == 0 {
		return false, nil
	}

	for _, row := range rows {
		grant := row.Str(0)
		// This regex hasn't been extensively tested but works in my local tests
		regex := fmt.Sprintf("GRANT ALL (.+) ON .%v.(.+) TO .%v.@.%v.", mu.Database, mu.Username, mu.Hostname)
		matched, err := regexp.MatchString(regex, grant)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

type MysqlDatabase struct {
	Name string
	Result
	MysqlConnInfo *MysqlConnInfo
}

func NewMysqlDatabase(name string) *MysqlDatabase {
	connInfo := &MysqlConnInfo{"root", "", "127.0.0.1", 3306}
	return &MysqlDatabase{Name: name, MysqlConnInfo: connInfo}
}

func (mydb *MysqlDatabase) Run() {
	db := mysql.New(
		"tcp",
		"",
		fmt.Sprintf("%v:%v", mydb.MysqlConnInfo.Host, mydb.MysqlConnInfo.Port),
		mydb.MysqlConnInfo.Username,
		mydb.MysqlConnInfo.Password,
		"mysql",
	)

	err := db.Connect()
	if err != nil {
		mydb.Result.Error = err.Error()
		return
	}

	exists, err := mydb.DatabaseExists(db)
	if err != nil {
		mydb.Result.Error = err.Error()
		return
	}

	if exists {
		mydb.Result.Success = true
		return
	}

	err = mydb.CreateDatabase(db)
	if err != nil {
		mydb.Result.Error = err.Error()
		return
	}

	mydb.Result.Success = true
	return
}

func (mydb *MysqlDatabase) CreateDatabase(db mysql.Conn) error {
	// IF NOT EXISTS isn't strictly necessary since we are checking
	// to make sure it doesn't exist first but it doesn't hurt.
	_, _, err := db.Query("CREATE DATABASE IF NOT EXISTS %v", mydb.Name)
	if err != nil {
		return err
	}
	mydb.Result.Changed = true
	return nil
}

func (mydb *MysqlDatabase) DatabaseExists(db mysql.Conn) (bool, error) {
	rows, _, err := db.Query(
		"SELECT count(*) AS dbs FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = '%v'",
		mydb.Name,
	)
	if err != nil {
		return false, err
	}

	if len(rows) != 1 {
		return false, errors.New("Unexpected number of rows in DatabaseExists()")
	}

	if rows[0].Int(0) > 0 {
		return true, nil
	}
	return false, nil
}
