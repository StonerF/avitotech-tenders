package save

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"avitotech/tenders/internal/lib/logger/sl"
	"avitotech/tenders/internal/storage"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

const (
	StatusOK    = "OK"
	StatusError = "Error"
)

type Request struct {
	Len   int     `json:"len"`
	Graph [][]int `json:"graph"`
}
type Response struct {
	Status string  `json:"status"`
	Error  string  `json:"error,omitempty"`
	Graph  [][]int `json:"graphs"`
}

type GraphSave interface {
	Save(Len int, Graph [][]int) (int64, error)
}

func New(log *slog.Logger, graphsaver GraphSave) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.http-server.handlers.save.New"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			render.JSON(w, r, Response{
				Status: StatusError,
				Error:  "empty request",
				Graph:  nil,
			})

			return
		}
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, Response{
				Status: StatusError,
				Error:  "failed to decode request",
				Graph:  nil,
			})

			return
		}
		log.Info("request body decoded", slog.Any("request", req))

		id, err := graphsaver.Save(req.Len, req.Graph)
		if errors.Is(err, storage.ErrGrisExist) {
			log.Info("graph already exist", slog.String("graph", strings.ReplaceAll(fmt.Sprintf("%d", req.Graph), " ", ",")), slog.Int("len", req.Len))

			render.JSON(w, r, Response{
				Status: StatusError,
				Error:  "graph already exist",
				Graph:  nil,
			})

			return
		}
		if err != nil {
			log.Error("failed to add graph", sl.Err(err))

			render.JSON(w, r, Response{
				Status: StatusError,
				Error:  "failed to add graph",
				Graph:  nil,
			})

			return
		}

		log.Info("graph added", slog.Int64("id", id))
		responseOK(w, r, req.Graph)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, graph [][]int) {
	render.JSON(w, r, Response{
		Status: StatusOK,
		Graph:  graph,
	})
}
