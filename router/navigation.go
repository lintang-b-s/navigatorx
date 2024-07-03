package router

import (
	"context"
	"errors"
	"lintang/coba_osm/alg"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type NavigationService interface {
	ShortestPath(ctx context.Context, srcLat, srcLon float64,
		dstLat float64, dstLon float64) (string, float64, bool, []alg.Coordinate, error)
}

type NavigationHandler struct {
	svc NavigationService
}

func NavigatorRouter(r *chi.Mux, svc NavigationService) {
	handler := &NavigationHandler{svc}

	r.Group(func(r chi.Router) {
		r.Route("/api/navigations", func(r chi.Router) {
			r.Post("/shortestPath", handler.shortestPath)
		})
	})
}

type SortestPathRequest struct {
	SrcLat float64 `json:"src_lat"`
	SrcLon float64 `json:"src_lon"`
	DstLat float64 `json:"dst_lat"`
	DstLon float64 `json:"dst_lon"`
}

func (s *SortestPathRequest) Bind(r *http.Request) error {
	if s.SrcLat == 0 || s.SrcLon == 0 || s.DstLat == 0 || s.DstLon == 0 {
		return errors.New("invalid request")
	}
	return nil
}

type ShortestPathResponse struct {
	Path  string           `json:"path"`
	Dist  float64          `json:"distance"`
	Found bool             `json:"found"`
	Route []alg.Coordinate `json:"route"`
}

func NewShortestPathResponse(path string, distance float64, route []alg.Coordinate, found bool) *ShortestPathResponse {

	return &ShortestPathResponse{
		Path:  path,
		Dist:  distance,
		Found: found,
		Route: route,
	}
}

func (h *NavigationHandler) shortestPath(w http.ResponseWriter, r *http.Request) {
	data := &SortestPathRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	p, dist, found, route, err := h.svc.ShortestPath(r.Context(), data.SrcLat, data.SrcLon, data.DstLat, data.DstLon)
	if err != nil {
		render.Render(w, r, ErrInternalServerError(errors.New("internal server error")))
		return
	}
	if !found {
		render.Render(w, r, ErrInvalidRequest(errors.New("node not found")))
		return
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, NewShortestPathResponse(p, dist, route, found))
}

type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInternalServerError(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 500,
		StatusText:     "Internal server error.",
		ErrorText:      err.Error(),
	}
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}
