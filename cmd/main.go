package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"syscall"
	"time"

	"avitotech/tenders/internal/config"
	"avitotech/tenders/internal/http-server/handlers/api/bids"
	"avitotech/tenders/internal/http-server/handlers/api/tenders"
	"avitotech/tenders/internal/lib/logger/sl"
	"avitotech/tenders/internal/lib/logger/slogpretty"
	"avitotech/tenders/internal/storage"
	"avitotech/tenders/internal/storage/postgres"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

const (
	envlocal             = "local"
	envDev               = "dev"
	envProd              = "prod"
	host                 = "rc1b-5xmqy6bq501kls4m.mdb.yandexcloud.net"
	port                 = 6432
	dbname               = "cnrprod1725724920-team-79197"
	user                 = "cnrprod1725724920-team-79197"
	password             = "cnrprod1725724920-team-79197"
	target_session_attrs = "read-write"
)

type ResponseError struct {
	Reason string `json:"reason"`
}

func main() {
	typeservices := map[string]struct{}{
		"Construction": struct{}{},
		"Delivery":     struct{}{},
		"Manufacture":  struct{}{},
	}
	// TODO : init config : cleanenv
	Cfg := config.MustLoad()

	fmt.Println(Cfg)

	// TODO : init logger : slog
	log := setupLogger("local")
	log.Info("starting service", slog.String("env", "local"))
	log.Debug("debug message are enabled")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=require target_session_attrs=%s",
		host, port, user, password, dbname, target_session_attrs)

	// TODO : init db : Postresql(sqlite)

	// TODO : init router : chi

	r := chi.NewRouter()
	// middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	/*r.Use(middleware.BasicAuth("/", map[string]string{
		"admin": "admin",
	})) */
	r.Get("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		storage, err1 := postgres.New(psqlInfo)
		if err1 != nil {
			log.Error("failed to init storage", sl.Err(err1))
			os.Exit(1)
		}
		storage.Ping()
		_, err := w.Write([]byte("ok"))
		if err != nil {
			log.Info("error response")
		}
		storage.Close()
		log.Info("otvet poluchen", slog.Any("req", w))
	})
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("ti loh ahahhhahhahhahhaah"))
		if err != nil {
			log.Info("error response")
		}
	})
	r.Post("/api/tenders/new", func(w http.ResponseWriter, r *http.Request) {
		req := tenders.Tender{}
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Info("error decode", slog.Any("err:", err))
		}
		//fmt.Println(req)
		tenobj, errten := tenders.Create(req)
		if errten != nil {
			log.Info("error db", slog.Any("err:", errten))
		}
		tenderid, errsend := tenobj.Send(psqlInfo)
		fmt.Println(tenderid)
		if errors.Is(errsend, sql.ErrNoRows) {
			w.WriteHeader(http.StatusUnauthorized)
			render.JSON(w, r, ResponseError{
				Reason: "username not authorized or not exist",
			})
			return
		}

		if errors.Is(errsend, storage.ErrGrisExist) {
			w.WriteHeader(http.StatusForbidden)
			render.JSON(w, r, ResponseError{
				Reason: "username have not roots for create tender for this organization",
			})
			return
		}
		if errors.Is(errsend, storage.Errnotfoundresp) {
			w.WriteHeader(http.StatusForbidden)
			render.JSON(w, r, ResponseError{
				Reason: "username have not roots for create tender for him organization",
			})
			return
		}

		resp := tenobj.GetbyTenderId(tenderid, psqlInfo)
		render.JSON(w, r, &resp)

	})
	r.Get("/api/tenders/my", func(w http.ResponseWriter, r *http.Request) {
		lim := r.URL.Query().Get("limit")
		offset := r.URL.Query().Get("offset")
		username := r.URL.Query().Get("username")
		id := ""
		db, _ := sql.Open("postgres", psqlInfo)
		stmt, _ := db.Prepare(`SELECT id FROM employee WHERE username = $1 `)
		err := stmt.QueryRow(username).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusUnauthorized)
			render.JSON(w, r, ResponseError{
				Reason: "username don't authorized or not exist",
			})
			return
		}
		defer db.Close()
		limnum, err := strconv.Atoi(lim)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, ResponseError{
				Reason: "invalid query string",
			})
			return
		}
		//fmt.Println("checkpoint")
		offnum, err := strconv.Atoi(offset)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, ResponseError{
				Reason: "invalid query string",
			})
			return
		}
		dop := ` where creatorusername = '` + username + `'`

		ans := tenders.Getmy(psqlInfo, dop, db)
		sort.Slice(ans, func(i, j int) bool {
			return ans[i].Name < ans[j].Name
		})
		if offnum >= len(ans) {
			ans = make([]tenders.TenderResponse, 0)
		} else if offnum+limnum > len(ans) {
			ans = ans[offnum:]
		} else if offnum == 0 && limnum < len(ans) {
			ans = ans[:limnum]
		} else {
			ans = ans[offnum : offnum+limnum]
		}
		render.JSON(w, r, ans)

	})
	/*r.Get("/api/tenders", func(w http.ResponseWriter, r *http.Request) {
		ans := tenders.Get(psqlInfo, "")
		sort.Slice(ans, func(i, j int) bool {
			return ans[i].Name < ans[j].Name
		})
		render.JSON(w, r, ans)
	}) */
	r.Get("/api/tenders", func(w http.ResponseWriter, r *http.Request) {
		lim := r.URL.Query().Get("limit")
		offset := r.URL.Query().Get("offset")
		/* if ser1 := r.URL.Query().Get("service_type"); ser1 != "" {
			ser_type = append(ser_type, ser1)
		}
		if ser2 := r.URL.Query().Get("service_type"); ser2 != "" {
			ser_type = append(ser_type, ser2)
		}
		if ser3 := r.URL.Query().Get("service_type"); ser3 != "" {
			ser_type = append(ser_type, ser3)
		}*/
		ser_type := r.URL.Query()["service_type"]
		limnum, err := strconv.Atoi(lim)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, ResponseError{
				Reason: "invalid query string",
			})
			return
		}
		fmt.Println("checkpoint")
		offnum, err := strconv.Atoi(offset)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, ResponseError{
				Reason: "invalid query string",
			})
			return
		}
		dop := ` `
		if len(ser_type) == 0 {
			dop = ""
		} else {

			for i := range ser_type {
				if _, ok := typeservices[ser_type[i]]; !ok {
					w.WriteHeader(http.StatusBadRequest)
					render.JSON(w, r, ResponseError{
						Reason: "invalid query string",
					})
					return
				}
				if ser_type[i] != "" {
					if i != 0 {
						dop += ` `
					}
					dop += `type=`
					dop += `'`
					dop += ser_type[i]
					dop += `'`
					dop += ` `
					if i != len(ser_type)-1 {
						dop += `OR`
					}
				}
				//fmt.Println(dop)

			}
		}
		db, err := sql.Open("postgres", psqlInfo)
		if err != nil {
			panic(err)
		}
		ans := tenders.Get(psqlInfo, dop, db)
		sort.Slice(ans, func(i, j int) bool {
			return ans[i].Name < ans[j].Name
		})
		if offnum >= len(ans) {
			ans = make([]tenders.TenderResponse, 0)
		} else if offnum+limnum > len(ans) {
			ans = ans[offnum:]
		} else if offnum == 0 && limnum < len(ans) {
			ans = ans[:limnum]
		} else {
			ans = ans[offnum : offnum+limnum]
		}
		render.JSON(w, r, ans)

	})
	r.Get("/api/tenders/{tenderId:[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$}/status", func(w http.ResponseWriter, r *http.Request) {
		tenderid := chi.URLParam(r, "tenderId")

		err := uuid.Validate(tenderid)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, ResponseError{
				Reason: "invalid query string",
			})
			return
		}
		username := r.URL.Query().Get("username")
		if username == "" {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, ResponseError{
				Reason: "invalid query string",
			})
			return
		}
		id := ""

		db, _ := sql.Open("postgres", psqlInfo)
		isexsists := tenders.Tenderidexist(tenderid, db)
		if !isexsists {
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, ResponseError{
				Reason: "tender not found",
			})
			return
		}
		stmt, _ := db.Prepare(`SELECT id FROM employee WHERE username = $1 `)
		err = stmt.QueryRow(username).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusUnauthorized)
			render.JSON(w, r, ResponseError{
				Reason: "username don't authorized or not exist",
			})
			return
		}
		defer db.Close()
		status, orgid, err := tenders.Getstatus(tenderid, psqlInfo, db)
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, ResponseError{
				Reason: "not found tenders",
			})
			return
		}
		if status == "Published" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Published"))
		} else {
			orgOutside, flag := tenders.ResponsibleClient(id, psqlInfo, db)
			if flag == false {
				w.WriteHeader(http.StatusForbidden)
				render.JSON(w, r, ResponseError{
					Reason: "user has not root",
				})
				return
			} else {
				if orgOutside != orgid {
					w.WriteHeader(http.StatusForbidden)
					render.JSON(w, r, ResponseError{
						Reason: "user has not root",
					})
					return
				} else {
					w.Write([]byte(status))
				}
			}

		}

	})
	r.Post("/api/bids/new", func(w http.ResponseWriter, r *http.Request) {
		db, err := sql.Open("postgres", psqlInfo)
		if err != nil {
			panic(err)
		}
		var Bid bids.Bids
		err = render.DecodeJSON(r.Body, &Bid)
		if Bid.AuthorType == "Orgnization" {
			IsEx := tenders.OrgIsExsits(Bid.AuthorId, db)
			if !IsEx {
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w, r, ResponseError{
					Reason: "organization doesn`t exist",
				})
				return
			}
			_, _, err := tenders.Getstatus(Bid.TenderId, psqlInfo, db)
			if errors.Is(err, sql.ErrNoRows) {
				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, ResponseError{
					Reason: "tenders not found",
				})
				return
			}
		} else {
			_, _, err := tenders.Getstatus(Bid.TenderId, psqlInfo, db)
			if errors.Is(err, sql.ErrNoRows) {
				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, ResponseError{
					Reason: "tenders not found",
				})
				return
			}
			iduser, flag := tenders.Validuser(Bid.AuthorId, db)
			fmt.Println(iduser)
			if !flag {
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w, r, ResponseError{
					Reason: "user doesn`t exist",
				})
				return
			}
			_, flag2 := tenders.ResponsibleClient(Bid.AuthorId, psqlInfo, db)
			if !flag2 {
				w.WriteHeader(http.StatusForbidden)
				render.JSON(w, r, ResponseError{
					Reason: "user dont have root",
				})
				return
			}
		}
		if err != nil {
			panic(err)
		}
		bidsreq, errr := bids.Send(db, Bid)
		if errr != nil {
			panic(errr)
		}
		resp := bids.GetbyBidsId(bidsreq, db)
		render.JSON(w, r, &resp)

	})
	r.Get("/api/bids/my", func(w http.ResponseWriter, r *http.Request) {
		lim := r.URL.Query().Get("limit")
		offset := r.URL.Query().Get("offset")
		username := r.URL.Query().Get("username")
		id := ""
		db, _ := sql.Open("postgres", psqlInfo)
		stmt, _ := db.Prepare(`SELECT id FROM employee WHERE username = $1 `)
		err := stmt.QueryRow(username).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusUnauthorized)
			render.JSON(w, r, ResponseError{
				Reason: "username don't authorized or not exist",
			})
			return
		}
		defer db.Close()
		limnum, err := strconv.Atoi(lim)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, ResponseError{
				Reason: "invalid query string",
			})
			return
		}
		//fmt.Println("checkpoint")
		offnum, err := strconv.Atoi(offset)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, ResponseError{
				Reason: "invalid query string",
			})
			return
		}

		ans := bids.GetmyBids(psqlInfo, db, id)
		sort.Slice(ans, func(i, j int) bool {
			return ans[i].Name < ans[j].Name
		})
		if offnum >= len(ans) {
			ans = make([]bids.BidsResponse, 0)
		} else if offnum+limnum > len(ans) {
			ans = ans[offnum:]
		} else if offnum == 0 && limnum < len(ans) {
			ans = ans[:limnum]
		} else {
			ans = ans[offnum : offnum+limnum]
		}
		render.JSON(w, r, ans)

	})
	r.Get("/api/bids/{tenderId:[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}}/list", func(w http.ResponseWriter, r *http.Request) {
		tenderid := chi.URLParam(r, "tenderId")

		err := uuid.Validate(tenderid)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, ResponseError{
				Reason: "invalid query string",
			})
			return
		}
		lim := r.URL.Query().Get("limit")
		offset := r.URL.Query().Get("offset")
		username := r.URL.Query().Get("username")
		id := ""
		db, _ := sql.Open("postgres", psqlInfo)
		isexsists := tenders.Tenderidexist(tenderid, db)
		if !isexsists {
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, ResponseError{
				Reason: "tender not found",
			})
			return
		}
		stmt, _ := db.Prepare(`SELECT id FROM employee WHERE username = $1 `)
		err = stmt.QueryRow(username).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusUnauthorized)
			render.JSON(w, r, ResponseError{
				Reason: "username don't authorized or not exist",
			})
			return
		}
		defer db.Close()
		limnum, err := strconv.Atoi(lim)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, ResponseError{
				Reason: "invalid query string",
			})
			return
		}
		//fmt.Println("checkpoint")
		offnum, err := strconv.Atoi(offset)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, ResponseError{
				Reason: "invalid query string",
			})
			return
		}

		_, isRes := tenders.ResponsibleClient(id, psqlInfo, db)
		fmt.Println(isRes)
		if !isRes {
			if len(bids.GetmyBidslim(db, id, limnum, offnum, tenderid)) == 0 {
				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, ResponseError{
					Reason: "bids not found",
				})
			}

			render.JSON(w, r, bids.GetmyBidslim(db, id, limnum, offnum, tenderid))
			return
		} else {
			if len(bids.GetmyBidslim(db, id, limnum, offnum, tenderid)) == 0 {
				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, ResponseError{
					Reason: "bids not found",
				})
			}
			render.JSON(w, r, bids.GetmyBidslimResp(db, id, limnum, offnum, tenderid))
			return
		}

	})
	r.Get("/api/bids/{bidsId:[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$}/status", func(w http.ResponseWriter, r *http.Request) {
		bidsid := chi.URLParam(r, "bidsId")

		err := uuid.Validate(bidsid)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, ResponseError{
				Reason: "invalid query string",
			})
			return
		}
		username := r.URL.Query().Get("username")
		if username == "" {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, ResponseError{
				Reason: "invalid query string",
			})
			return
		}
		id := ""

		db, _ := sql.Open("postgres", psqlInfo)
		isexsists := bids.Bidsidexist(bidsid, db)
		if !isexsists {
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, ResponseError{
				Reason: "bids not found",
			})
			return
		}
		stmt, _ := db.Prepare(`SELECT id FROM employee WHERE username = $1 `)
		err = stmt.QueryRow(username).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusUnauthorized)
			render.JSON(w, r, ResponseError{
				Reason: "username don't authorized or not exist",
			})
			return
		}
		defer db.Close()
		_, isresp := tenders.ResponsibleClient(id, psqlInfo, db)
		bidsobj := bids.GetbyBidsId(bidsid, db)
		if !isresp {

			if bidsobj.AuthorId == id {
				w.Write([]byte(bidsobj.Status))
				return
			} else {
				w.WriteHeader(http.StatusForbidden)
				render.JSON(w, r, ResponseError{
					Reason: "user dont have root",
				})
				return
			}

		} else {
			w.Write([]byte(bidsobj.Status))
			return
		}

	})

	log.Info("starting server", slog.String("address", Cfg.Address))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	srv := &http.Server{
		Addr:         "0.0.0.0:8080",
		Handler:      r,
		ReadTimeout:  Cfg.HttpServer.Timeout,
		WriteTimeout: Cfg.HttpServer.Timeout,
		IdleTimeout:  Cfg.HttpServer.Idle_timeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("failed to start server")
		}
	}()

	log.Info("server started")

	<-done
	log.Info("stopping server")

	// TODO: move timeout to config
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", sl.Err(err))

		return
	}

	// TODO: close storage

	log.Info("server stopped")

	// TODO : run server
	/*r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	r.Get("/loh", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("nu ti loh konesno"))
		fmt.Println("Serving", " ", r.URL, " ", r.Host)

	})

	http.ListenAndServe(":3000", r)
	*/

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envlocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default: // If env config is invalid, set prod settings by default due to security
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
