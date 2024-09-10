package postgres

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func New(str string) (*Storage, error) {
	const op = "storage.postgres.New"
	connStr := str
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	return &Storage{db: db}, nil
}
func (str *Storage) Ping() {
	err := str.db.Ping()
	if err != nil {
		log.Fatal("don't ping")
	}
}

/*
func (s *Storage) IsExist(Len int, Graph [][]int) (bool, error) {
	const op = "storage.postgres.IsExist"
	lenstr := strconv.Itoa(Len)
	var find_ind int64
	req := "ARRAY" + strings.ReplaceAll(fmt.Sprintf("%d", Graph), " ", ",")
	errfind := s.db.QueryRow(`SELECT id FROM graphs WHERE edges =` + req + `AND len =` + lenstr + `;`).Scan(&find_ind)
	if errfind != nil {
		return false, fmt.Errorf("%s: %w", op, errfind)
	}
	if find_ind != 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func (s *Storage) Save(Len int, Graph [][]int) (int64, error) {
	const op = "storage.postgres.Save"
	req := "ARRAY" + strings.ReplaceAll(fmt.Sprintf("%d", Graph), " ", ",")
	lenstr := strconv.Itoa(Len)
	fmt.Println(req)

	stmt, err := s.db.Prepare(`INSERT INTO graphs(len,edges) VALUES(` + lenstr + `,` + req + `) RETURNING id;`)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()
	if ok, _ := s.IsExist(Len, Graph); !ok {
		var userid int64
		errQuery := stmt.QueryRow().Scan(&userid)
		if errQuery != nil {
			return 0, fmt.Errorf("%s: %w", op, errQuery)
		}

		return userid, nil
	} else {
		return 0, fmt.Errorf("%s:%w", op, storage.ErrGrisExist)
	}
}

// TODO: do GET when will be more code
//func (s Storage) Get()
*/
