package main

import (
	_ "github.com/bmizerany/pq"

	"crypto/md5"
	"database/sql"
	"fmt"
	"io"
	"strings"
)

type PgUser struct {
	Username string
	Password string
	Result
	PgConnInfo
}

func NewPgUser(username, password string) *PgUser {
	return &PgUser{Username: username, Password: password}
}

func (pg *PgUser) Run() {
	dbConnString := pg.PgConnInfo.GetConnString()
	db, err := sql.Open("postgres", dbConnString)
	if err != nil {
		pg.Result.Error = err.Error()
		return
	}
	defer db.Close()

	userExists, correctPass, err := pg.UserCheck(db)
	if err != nil {
		pg.Result.Error = err.Error()
		return
	}

	if userExists != true || correctPass != true {
		var q string
		if userExists != true {
			q = fmt.Sprintf("CREATE USER %v ENCRYPTED PASSWORD '%v'", pg.Username, pg.ComputePassword())
		} else if correctPass != true {
			q = fmt.Sprintf("ALTER USER %v ENCRYPTED PASSWORD '%v'", pg.Username, pg.ComputePassword())
		}

		_, err = db.Exec(q)
		if err != nil {
			pg.Result.Error = err.Error()
			return
		}
		pg.Result.Changed = true
	}

	pg.Result.Success = true
	return
}

func (pg *PgUser) UserCheck(db *sql.DB) (userExists, correctPass bool, err error) {
	q := "SELECT passwd FROM pg_shadow WHERE usename = $1 "
	rows, err := db.Query(q, pg.Username)
	if err != nil {
		return false, false, err
	}
	for rows.Next() {
		userExists = true // We got at least one row so we know the user exists
		var p string
		err = rows.Scan(&p)
		if err != nil {
			return userExists, correctPass, err
		}
		if p == pg.ComputePassword() {
			correctPass = true
			return userExists, correctPass, nil
		}
	}
	return userExists, correctPass, nil
}

func (pg *PgUser) ComputePassword() string {
	// password is already computed so just return it
	if strings.HasPrefix(pg.Password, "md5") {
		return pg.Password
	}

	s := pg.Password + pg.Username
	h := md5.New()
	io.WriteString(h, s)
	return fmt.Sprintf("md5%x", h.Sum(nil))
}

type PgDatabase struct {
	Name  string
	Owner string
	Result
	PgConnInfo
}

func NewPgDatabase(name, owner string) *PgDatabase {
	return &PgDatabase{Name: name, Owner: owner}
}

func (pgd *PgDatabase) Run() {
	dbConnString := pgd.PgConnInfo.GetConnString()
	db, err := sql.Open("postgres", dbConnString)
	if err != nil {
		pgd.Result.Error = err.Error()
		return
	}
	defer db.Close()

	dbExists, correctOwner, err := pgd.DatabaseCheck(db)
	if err != nil {
		pgd.Result.Error = err.Error()
		return
	}

	if dbExists != true {
		q := fmt.Sprintf("CREATE DATABASE %v WITH OWNER = %v", pgd.Name, pgd.Owner)
		_, err = db.Exec(q)
		if err != nil {
			pgd.Result.Error = err.Error()
			return
		}
		pgd.Result.Changed = true
	} else if correctOwner != true {
		q := fmt.Sprintf("ALTER DATABASE %v OWNER TO %v", pgd.Name, pgd.Owner)
		_, err = db.Exec(q)
		if err != nil {
			pgd.Result.Error = err.Error()
			return
		}
		pgd.Result.Changed = true
	}

	pgd.Result.Success = true
	return
}

func (pgd *PgDatabase) DatabaseCheck(db *sql.DB) (dbExists, correctOwner bool, err error) {
	q := `SELECT pg_user.usename from pg_database
        LEFT JOIN pg_user ON pg_database.datdba = pg_user.usesysid
        WHERE datname=$1`
	rows, err := db.Query(q, pgd.Name)
	if err != nil {
		return
	}
	for rows.Next() {
		dbExists = true // We got at least one row so we know the db exists
		var owner string
		err = rows.Scan(&owner)
		if err != nil {
			return
		}
		if owner == pgd.Owner {
			correctOwner = true
			return
		}
	}
	// SELECT * from pg_database
	// LEFT JOIN pg_user ON pg_database.datdba = pg_user.usesysid
	// WHERE datname='shepherd';

	// ALTER DATABASE {name} OWNER TO {owner};
	return
}

type PgConnInfo struct {
	Username string
	Password string
	Host     string
	Port     int64
	Ssl      bool
}

func (ci *PgConnInfo) GetConnString() string {
	dbConnString := "dbname=postgres"
	if ci.Username != "" {
		dbConnString += fmt.Sprintf(" user=%v", ci.Username)
	}
	if ci.Password != "" {
		dbConnString += fmt.Sprintf(" password=%v", ci.Password)
	}
	if ci.Host != "" {
		dbConnString += fmt.Sprintf(" host=%v", ci.Host)
	}
	if ci.Port != 0 {
		dbConnString += fmt.Sprintf(" port=%v", ci.Port)
	}
	if ci.Ssl {
		dbConnString += " sslmode=require"
	} else {
		dbConnString += " sslmode=disable"
	}
	return dbConnString
}
