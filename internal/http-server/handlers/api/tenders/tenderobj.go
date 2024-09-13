package tenders

import (
	"avitotech/tenders/internal/storage"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
)

type tenderObj struct {
	ten Tender
	db  *sql.DB
}

func Create(t Tender) (*tenderObj, error) {

	return &tenderObj{ten: Tender{
		Name:            t.Name,
		Description:     t.Description,
		ServiceType:     t.ServiceType,
		Status:          t.Status,
		OrganizationId:  t.OrganizationId,
		CreatorUsername: t.CreatorUsername,
		Version:         1}}, nil
}
func (ten *tenderObj) Getnameuser() string {
	return ten.ten.CreatorUsername
}
func (ten *tenderObj) Send(path string) (string, error) {
	var err1 error
	ten.db, err1 = sql.Open("postgres", path)
	defer ten.db.Close()
	if err1 != nil {
		return "", fmt.Errorf("%w", err1)
	}
	//defer ten.db.Close()
	/*stmt0, err := ten.db.Prepare(`
		CREATE TYPE IF NOT EXISTS servicetype AS ENUM (
	    'Construction',
	    'Delivery',
	    'Manufacture'
	);`)
		if err != nil {
			log.Fatal(err)
			return "", fmt.Errorf("%w", err)
		}
		_, errexec := stmt0.Exec()
		if errexec != nil {
			log.Fatal(errexec)
			return "", fmt.Errorf("%w", errexec)
		}
		stmt1, err := ten.db.Prepare(`CREATE TYPE IF NOT EXISTS status_type AS ENUM (
	    'Created',
	    'Published',
	    'Closed'
	);`)
		if err != nil {
			log.Fatal(err)
			return "", fmt.Errorf("%w", err)
		}
		_, err = stmt1.Exec()
		if err != nil {
			log.Fatal(err)
			return "", fmt.Errorf("%w", err)
		}
	*/
	stmtcheck, errcheck := ten.db.Prepare(`SELECT id FROM employee WHERE username = $1 `)
	if errcheck != nil {
		panic(errcheck)
	}
	id := ""
	row := stmtcheck.QueryRow(ten.ten.CreatorUsername).Scan(&id)
	if errors.Is(row, sql.ErrNoRows) {
		return "", sql.ErrNoRows
	}
	stmtcheck, errcheck = ten.db.Prepare(`SELECT organization_id FROM organization_responsible WHERE user_id = $1 `)
	if errcheck != nil {
		panic(errcheck)
	}
	org_id := ""
	row = stmtcheck.QueryRow(id).Scan(&org_id)
	if errors.Is(row, sql.ErrNoRows) {
		return "", storage.Errnotfoundresp
	}
	fmt.Println(org_id)
	if ten.ten.OrganizationId != org_id {
		return "", storage.ErrGrisExist
	}

	stmt2, err := ten.db.Prepare(`CREATE TABLE IF NOT EXISTS tenders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
	description TEXT NOT NULL,
    type servicetype,
    status  status_type,
    organizationid UUID,
    creatorusername VARCHAR(50),
	version INT,
	createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`)
	if err != nil {
		log.Fatal(err, 1)
		return "", fmt.Errorf("%w", err)
	}
	_, err = stmt2.Exec()
	if err != nil {
		log.Fatal(err, 2)
		return "", fmt.Errorf("%w", err)
	}

	stmt, errpr := ten.db.Prepare(`INSERT INTO tenders(name,description,type,status,organizationid,creatorusername,version) VALUES(` + `'` + ten.ten.Name + `'` + `,` + `'` + ten.ten.Description + `'` + `,` + `'` + ten.ten.ServiceType + `'` + `,` + `'` + "Created" + `'` + `,` + `'` + ten.ten.OrganizationId + `'` + `,` + `'` + ten.ten.CreatorUsername + `'` + `,` + strconv.Itoa(1) + `) RETURNING id;`)
	if errpr != nil {
		log.Fatal(errpr, 3)
		//log.Fatal("err prepare")
		return "", fmt.Errorf("%w", errpr)
	}
	var userid string
	errQuery := stmt.QueryRow().Scan(&userid)
	if errQuery != nil {
		log.Fatal(err)
		//log.Fatal("err query")
		return "", fmt.Errorf("%w", errQuery)
	}

	return userid, nil

}
func (ten *tenderObj) Get(path string, dop string) []Tender {
	var err1 error

	ten.db, err1 = sql.Open("postgres", path)
	if err1 != nil {
		return nil
	}
	defer ten.db.Close()
	otvet := `select * from tenders where`
	otvet += dop
	//defer ten.db.Close()
	rows, err := ten.db.Query(otvet)
	if err != nil {
		panic(err)
	}
	tenders := []Tender{}
	for rows.Next() {
		p := Tender{}
		err := rows.Scan(&p.Id, &p.Name, &p.Description, &p.ServiceType, &p.Status, &p.OrganizationId, &p.CreatorUsername, &p.Version, &p.CreatedAt)
		if err != nil {
			fmt.Println(err)
			continue
		}
		tenders = append(tenders, p)
	}
	defer rows.Close()
	return tenders

}
func (ten *tenderObj) GetbyTenderId(id string, path string) Tender {
	var err1 error

	ten.db, err1 = sql.Open("postgres", path)
	if err1 != nil {
		log.Fatal("don't conn Getby ten id")
	}

	defer ten.db.Close()
	rows, err := ten.db.Query(`select * from tenders where id =` + `'` + id + `'` + `;`)
	if err != nil {
		panic(err)
	}
	tenders := Tender{}
	for rows.Next() {
		p := Tender{}
		err := rows.Scan(&p.Id, &p.Name, &p.Description, &p.ServiceType, &p.Status, &p.OrganizationId, &p.CreatorUsername, &p.Version, &p.CreatedAt)
		if err != nil {
			fmt.Println(err)
			continue
		}
		tenders = p
	}
	defer rows.Close()
	return tenders

}
func Get(path string, dop string, db *sql.DB) []TenderResponse {

	otvet := `select * from tenders`
	if dop == `` {
		otvet += ` WHERE status='Published' `
	} else {
		otvet += ` WHERE `
		otvet += `(`
		otvet += dop
		otvet += `)`
		otvet += ` AND `
		otvet += `(status='Published')`
	}
	fmt.Println(otvet)
	//defer ten.db.Close()
	rows, err := db.Query(otvet)
	if err != nil {
		panic(err)
	}
	tenders := []TenderResponse{}
	var username string
	var orgid string
	for rows.Next() {
		p := TenderResponse{}
		err := rows.Scan(&p.Id, &p.Name, &p.Description, &p.ServiceType, &p.Status, &orgid, &username, &p.Version, &p.CreatedAt)
		if err != nil {
			fmt.Println(err)
			continue
		}
		tenders = append(tenders, p)
	}
	defer rows.Close()
	return tenders

}
func Getmy(path string, dop string, db *sql.DB) []TenderResponse {
	otvet := `select * from tenders `
	otvet += dop

	fmt.Println(otvet)
	//defer ten.db.Close()
	rows, err := db.Query(otvet)
	if err != nil {
		panic(err)
	}
	tenders := []TenderResponse{}
	var username string
	var orgid string
	for rows.Next() {
		p := TenderResponse{}
		err := rows.Scan(&p.Id, &p.Name, &p.Description, &p.ServiceType, &p.Status, &orgid, &username, &p.Version, &p.CreatedAt)
		if err != nil {
			fmt.Println(err)
			continue
		}
		tenders = append(tenders, p)
	}
	defer rows.Close()
	return tenders

}
func Getstatus(tenderid string, path string, db *sql.DB) (string, string, error) {

	stmt, err := db.Prepare(`SELECT status,organizationid FROM tenders WHERE id =$1`)
	if err != nil {
		panic(err)
	}
	Stat := ""
	org_id := ""

	//defer ten.db.Close()
	row := stmt.QueryRow(tenderid).Scan(&Stat, &org_id)
	if errors.Is(row, sql.ErrNoRows) {
		return "", "", sql.ErrNoRows
	}
	return Stat, org_id, nil
}

func ResponsibleClient(id string, path string, db *sql.DB) (string, bool) {
	fmt.Println(id)
	stmt, err := db.Prepare(`SELECT organization_id FROM organization_responsible WHERE user_id = $1`)
	if err != nil {
		panic(stmt)
	}
	org_id := ""

	//defer ten.db.Close()
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
