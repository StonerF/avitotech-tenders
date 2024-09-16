package auth

import (
	"database/sql"
	"errors"
	"fmt"
)

func ResponsibleClient(id string, path string, db *sql.DB) (string, bool) {
	fmt.Println(id)
	stmt, err := db.Prepare(`SELECT organization_id FROM organization_responsible WHERE user_id = $1`)
	if err != nil {
		panic(stmt)
	}
	org_id := ""
	row := stmt.QueryRow(id).Scan(&org_id)
	if errors.Is(row, sql.ErrNoRows) {
		return "", false
	}
	return org_id, true
}
func Validuser(user string, db *sql.DB) (string, bool) {
	stmtcheck, errcheck := db.Prepare(`SELECT username FROM employee WHERE  id = $1 `)
	if errcheck != nil {
		panic(errcheck)
	}
	id := ""
	row := stmtcheck.QueryRow(user).Scan(&id)
	if errors.Is(row, sql.ErrNoRows) {
		return "", false
	}
	return id, true
}
func OrgIsExsits(org_id string, db *sql.DB) bool {
	stmtcheck, errcheck := db.Prepare(`SELECT name FROM organization WHERE id = $1 `)
	if errcheck != nil {
		panic(errcheck)
	}
	id := ""
	row := stmtcheck.QueryRow(org_id).Scan(&id)
	return !errors.Is(row, sql.ErrNoRows)
}
func Tenderidexist(tenderid string, db *sql.DB) bool {
	stmtcheck, errcheck := db.Prepare(`SELECT name FROM tenders WHERE id = $1 `)
	if errcheck != nil {
		panic(errcheck)
	}
	id := ""
	row := stmtcheck.QueryRow(tenderid).Scan(&id)
	return !errors.Is(row, sql.ErrNoRows)
}
func OrgFromTender(tenderid string, db *sql.DB) (string, bool) {
	stmtcheck, errcheck := db.Prepare(`SELECT organizationid FROM tenders WHERE id = $1 `)
	if errcheck != nil {
		panic(errcheck)
	}
	id := ""
	row := stmtcheck.QueryRow(tenderid).Scan(&id)
	return id, !errors.Is(row, sql.ErrNoRows)
}
