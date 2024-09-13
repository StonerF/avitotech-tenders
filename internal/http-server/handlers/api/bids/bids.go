package bids

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
)

type Bids struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	TenderId    string `json:"tenderId"`
	AuthorType  string `json:"authorType"`
	AuthorId    string `json:"authorId"`
	Version     int    `json:"version"`
	CreatedAt   string `json:"createdAt"`
}
type BidsResponse struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	AuthorType string `json:"authorType"`
	AuthorId   string `json:"authorId"`
	Version    int    `json:"version"`
	CreatedAt  string `json:"createdAt"`
}

func Send(db *sql.DB, bid Bids) (string, error) {

	stmt0, err := db.Prepare(`
		CREATE TYPE authortype AS ENUM(
	    'Organization',
	    'User'
	);`)
	if err != nil {

	}
	_, err = stmt0.Exec()
	if err != nil {

	}
	stmt1, err := db.Prepare(`CREATE TYPE status_typebid AS ENUM(
	    'Created',
	    'Published',
	    'Canceled'
	);`)
	if err != nil {

	}

	_, err = stmt1.Exec()
	if err != nil {

	}

	stmt2, err := db.Prepare(`CREATE TABLE IF NOT EXISTS bids (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
	description TEXT NOT NULL,
    status  status_typebid,
	tenderid UUID,
	type authortype,
    authorid UUID,
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

	stmt, errpr := db.Prepare(`INSERT INTO bids(name,description,status,tenderid,type,authorid,version) VALUES(` + `'` + bid.Name + `'` + `,` + `'` + bid.Description + `'` + `,` + `'` + "Created" + `'` + `,` + `'` + bid.TenderId + `'` + `,` + `'` + bid.AuthorType + `'` + `,` + `'` + bid.AuthorId + `'` + `,` + strconv.Itoa(1) + `) RETURNING id;`)
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
func GetbyBidsId(id string, db *sql.DB) BidsResponse {
	rows, err := db.Query(`select * from bids where id =` + `'` + id + `'` + `;`)
	if err != nil {
		panic(err)
	}
	tenders := BidsResponse{}
	desc := ""
	tenderid := ""
	for rows.Next() {
		p := BidsResponse{}
		err := rows.Scan(&p.Id, &p.Name, &desc, &p.Status, &tenderid, &p.AuthorType, &p.AuthorId, &p.Version, &p.CreatedAt)
		if err != nil {
			fmt.Println(err)
			continue
		}
		tenders = p
	}
	defer rows.Close()
	return tenders

}
func GetmyBids(dop string, db *sql.DB, userid string) []BidsResponse {

	otvet := `select * from bids Where (authorid = $1) AND (type = 'User')`

	fmt.Println(otvet)
	rows, err := db.Query(otvet, userid)
	if err != nil {
		panic(err)
	}
	tenders := []BidsResponse{}
	desc := ""
	tenderid := ""
	for rows.Next() {
		p := BidsResponse{}
		err := rows.Scan(&p.Id, &p.Name, &desc, &p.Status, &tenderid, &p.AuthorType, &p.AuthorId, &p.Version, &p.CreatedAt)
		if err != nil {
			fmt.Println(err)
			continue
		}
		tenders = append(tenders, p)
	}
	defer rows.Close()
	return tenders

}
func GetmyBidslim(db *sql.DB, userid string, limit int, offset int, tenderid string) []BidsResponse {

	otvet := `select * from bids Where (authorid = $1) AND (type = 'User') AND (tenderid = $4) Order By name LIMIT $2 OFFSET $3`

	fmt.Println(otvet)
	rows, err := db.Query(otvet, userid, limit, offset, tenderid)
	if err != nil {
		panic(err)
	}
	tenders := []BidsResponse{}
	desc := ""
	tenderidle := ""
	for rows.Next() {
		p := BidsResponse{}
		err := rows.Scan(&p.Id, &p.Name, &desc, &p.Status, &tenderidle, &p.AuthorType, &p.AuthorId, &p.Version, &p.CreatedAt)
		if err != nil {
			fmt.Println(err)
			continue
		}
		tenders = append(tenders, p)
	}
	defer rows.Close()
	return tenders

}
func GetmyBidslimResp(db *sql.DB, userid string, limit int, offset int, tenderid string) []BidsResponse {

	otvet := `select * from bids where tenderid=$1  Order By name LIMIT $2 OFFSET $3`

	fmt.Println(otvet)
	rows, err := db.Query(otvet, tenderid, limit, offset)
	if err != nil {
		panic(err)
	}
	tenders := []BidsResponse{}
	desc := ""
	tenderidle := ""
	for rows.Next() {
		p := BidsResponse{}
		err := rows.Scan(&p.Id, &p.Name, &desc, &p.Status, &tenderidle, &p.AuthorType, &p.AuthorId, &p.Version, &p.CreatedAt)
		if err != nil {
			fmt.Println(err)
			continue
		}
		tenders = append(tenders, p)
	}
	defer rows.Close()
	return tenders

}
func Bidsidexist(bidsid string, db *sql.DB) bool {
	stmtcheck, errcheck := db.Prepare(`SELECT name FROM bids WHERE id = $1 `)
	if errcheck != nil {
		panic(errcheck)
	}
	id := ""
	row := stmtcheck.QueryRow(bidsid).Scan(&id)
	return !errors.Is(row, sql.ErrNoRows)
}
