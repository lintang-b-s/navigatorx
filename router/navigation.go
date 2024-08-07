package router

import (
	"context"
	"errors"
	"fmt"
	"lintang/navigatorx/alg"
	"lintang/navigatorx/domain"
	"lintang/navigatorx/util"
	"math"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
)

type NavigationService interface {
	ShortestPathETA(ctx context.Context, srcLat, srcLon float64,
		dstLat float64, dstLon float64) (string, float64, []alg.Navigation, bool, []alg.Coordinate, float64, bool, error)

	ShortestPathAlternativeStreetETA(ctx context.Context, srcLat, srcLon float64,
		alternativeStreetLat float64, alternativeStreetLon float64,
		dstLat float64, dstLon float64) (string, float64, []alg.Navigation, bool, []alg.Coordinate, float64, bool, error)

	// ShortestPathETACH(ctx context.Context, srcLat, srcLon float64,
	// 	dstLat float64, dstLon float64) (string, float64, []alg.Navigation, bool, []alg.Coordinate, float64, error)
	ShortestPathETACH(ctx context.Context, srcLat, srcLon float64,
		dstLat float64, dstLon float64) (string, []alg.Navigation, []alg.Coordinate, float64, float64, error)

	HiddenMarkovModelMapMatching(ctx context.Context, gps []alg.Coordinate) (string, []alg.CHNode2, error)
}

type NavigationHandler struct {
	svc NavigationService
}

func NavigatorRouter(r *chi.Mux, svc NavigationService) {
	handler := &NavigationHandler{svc}

	r.Group(func(r chi.Router) {
		r.Route("/api/navigations", func(r chi.Router) {
			r.Post("/shortestPath", handler.shortestPathETA)
			r.Post("/shortestPathAlternativeStreet", handler.shortestPathAlternativeStreetETA)
			r.Post("/shortestPathCH", handler.shortestPathETACH)
			r.Post("/mapMatching", handler.HiddenMarkovModelMapMatching) 
		})
	})
}

type SortestPathRequest struct {
	SrcLat float64 `json:"src_lat" validate:"required,lt=90,gt=-90"`
	SrcLon float64 `json:"src_lon" validate:"required,lt=180,gt=-180"`
	DstLat float64 `json:"dst_lat" validate:"required,lt=90,gt=-90"`
	DstLon float64 `json:"dst_lon" validate:"required,lt=180,gt=-180"`
}

func (s *SortestPathRequest) Bind(r *http.Request) error {
	if s.SrcLat == 0 || s.SrcLon == 0 || s.DstLat == 0 || s.DstLon == 0 {
		return errors.New("invalid request")
	}
	return nil
}

type SortestPathAlternativeStreetRequest struct {
	SrcLat               float64 `json:"src_lat" validate:"required,lt=90,gt=-90"`
	SrcLon               float64 `json:"src_lon" validate:"required,lt=180,gt=-180"`
	StreetAlternativeLat float64 `json:"street_alternative_lat" validate:"required,lt=90,gt=-90"`
	StreetAlternativeLon float64 `json:"street_alternative_lon" validate:"required,lt=180,gt=-180"`
	DstLat               float64 `json:"dst_lat" validate:"required,lt=90,gt=-90"`
	DstLon               float64 `json:"dst_lon" validate:"required,lt=180,gt=-180"`
}

func (s *SortestPathAlternativeStreetRequest) Bind(r *http.Request) error {
	if s.SrcLat == 0 || s.SrcLon == 0 || s.StreetAlternativeLat == 0 || s.StreetAlternativeLon == 0 || s.DstLat == 0 || s.DstLon == 0 {
		return errors.New("invalid request")
	}
	return nil
}

type ShortestPathResponse struct {
	Path        string           `json:"path"`
	Dist        float64          `json:"distance,omitempty"`
	ETA         float64          `json:"ETA"`
	Navigations []alg.Navigation `json:"navigations"`
	Found       bool             `json:"found"`
	Route       []alg.Coordinate `json:"route,omitempty"`
	Alg         string           `json:"algorithm"`
}

func NewShortestPathResponse(path string, distance float64, navs []alg.Navigation, eta float64, route []alg.Coordinate, found bool, isCH bool) *ShortestPathResponse {

	var alg string
	if isCH {
		alg = "Contraction Hieararchies + Bidirectional Dijkstra"
	} else {
		alg = "A* Algorithm"
	}
	return &ShortestPathResponse{
		Path:        path,
		Dist:        util.RoundFloat(distance, 2),
		ETA:         util.RoundFloat(eta, 2),
		Navigations: navs,
		Found:       found,
		Alg:         alg,
	}
}

func (h *NavigationHandler) shortestPathETA(w http.ResponseWriter, r *http.Request) {
	data := &SortestPathRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	validate := validator.New()
	if err := validate.Struct(*data); err != nil {
		english := en.New()
		uni := ut.New(english, english)
		trans, _ := uni.GetTranslator("en")
		_ = enTranslations.RegisterDefaultTranslations(validate, trans)
		vv := translateError(err, trans)
		render.Render(w, r, ErrValidation(err, vv))
		return
	}

	p, dist, n, found, route, eta, isCH, err := h.svc.ShortestPathETA(r.Context(), data.SrcLat, data.SrcLon, data.DstLat, data.DstLon)
	if err != nil {
		if !found {
			render.Render(w, r, ErrInvalidRequest(errors.New("node not found")))
			return
		}
		render.Render(w, r, ErrInternalServerErrorRend(errors.New("internal server error")))
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, NewShortestPathResponse(p, dist, n, eta, route, found, isCH))
}

func (h *NavigationHandler) shortestPathAlternativeStreetETA(w http.ResponseWriter, r *http.Request) {
	data := &SortestPathAlternativeStreetRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	validate := validator.New()
	if err := validate.Struct(*data); err != nil {
		english := en.New()
		uni := ut.New(english, english)
		trans, _ := uni.GetTranslator("en")
		_ = enTranslations.RegisterDefaultTranslations(validate, trans)
		vv := translateError(err, trans)
		render.Render(w, r, ErrValidation(err, vv))
		return
	}

	p, dist, n, found, route, eta, isCH, err := h.svc.ShortestPathAlternativeStreetETA(r.Context(), data.SrcLat, data.SrcLon, data.StreetAlternativeLat, data.StreetAlternativeLon,
		data.DstLat, data.DstLon)
	if err != nil {
		if !found {
			render.Render(w, r, ErrInvalidRequest(errors.New("node not found")))
			return
		}
		render.Render(w, r, ErrInternalServerErrorRend(errors.New("internal server error")))
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, NewShortestPathResponse(p, dist, n, eta, route, found, isCH))
}

func (h *NavigationHandler) shortestPathETACH(w http.ResponseWriter, r *http.Request) {
	data := &SortestPathRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	validate := validator.New()
	if err := validate.Struct(*data); err != nil {
		english := en.New()
		uni := ut.New(english, english)
		trans, _ := uni.GetTranslator("en")
		_ = enTranslations.RegisterDefaultTranslations(validate, trans)
		vv := translateError(err, trans)
		render.Render(w, r, ErrValidation(err, vv))
		return
	}

	p, n, route, eta, dist, err := h.svc.ShortestPathETACH(r.Context(), data.SrcLat, data.SrcLon, data.DstLat, data.DstLon)
	if err != nil {
		// if !found {
		// 	render.Render(w, r, ErrInvalidRequest(errors.New("node not found")))
		// 	return
		// }
		render.Render(w, r, ErrInternalServerErrorRend(errors.New("internal server error")))
		return
	}
	found := false
	if eta != math.MaxFloat64 {
		found = true
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, NewShortestPathResponse(p, dist, n, eta, route, found, true))
}

type MapMatchingRequest struct {
	Coordinates []Coord `json:"coordinates" validate:"required,dive"`
}

type Coord struct {
	Lat float64 `json:"lat" validate:"required,lt=90,gt=-90"`
	Lon float64 `json:"lon" validate:"required,lt=180,gt=-180"`
}

func (s *MapMatchingRequest) Bind(r *http.Request) error {
	if len(s.Coordinates) == 0 {
		return errors.New("invalid request")
	}
	return nil
}

type MapMatchingResponse struct {
	Path        string  `json:"path"`
	Coordinates []Coord `json:"coordinates"`
}

func RenderMapMatchingResponse(path string, coords []alg.CHNode2) *MapMatchingResponse {
	coordsResp := []Coord{}
	for _, c := range coords {
		coordsResp = append(coordsResp, Coord{
			Lat: c.Lat,
			Lon: c.Lon,
		})
	}

	return &MapMatchingResponse{
		Path:        path,
		Coordinates: coordsResp,
	}
}

func (h *NavigationHandler) HiddenMarkovModelMapMatching(w http.ResponseWriter, r *http.Request) {
	data := &MapMatchingRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	validate := validator.New()
	if err := validate.Struct(*data); err != nil {
		english := en.New()
		uni := ut.New(english, english)
		trans, _ := uni.GetTranslator("en")
		_ = enTranslations.RegisterDefaultTranslations(validate, trans)
		vv := translateError(err, trans)
		render.Render(w, r, ErrValidation(err, vv))
		return
	}

	coords := []alg.Coordinate{}
	for _, c := range data.Coordinates {
		coords = append(coords, alg.Coordinate{
			Lat: c.Lat,
			Lon: c.Lon,
		})
	}
	p, pNode, err := h.svc.HiddenMarkovModelMapMatching(r.Context(), coords)
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorRend(errors.New("internal server error")))

		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, RenderMapMatchingResponse(p, pNode))
}

type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText    string   `json:"status"`          // user-level status message
	AppCode       int64    `json:"code,omitempty"`  // application-specific error code
	ErrorText     string   `json:"error,omitempty"` // application-level error message, for debugging
	ErrValidation []string `json:"validation,omitempty"`
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInternalServerErrorRend(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 500,
		StatusText:     "Internal server error.",
		ErrorText:      err.Error(),
	}
}

func ErrValidation(err error, errV []error) render.Renderer {
	vv := []string{}
	for _, v := range errV {
		vv = append(vv, v.Error())
	}
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
		ErrValidation:  vv,
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

func ErrChi(err error) render.Renderer {
	statusText := ""
	switch getStatusCode(err) {
	case http.StatusNotFound:
		statusText = "Resource not found."
	case http.StatusInternalServerError:
		statusText = "Internal server error."
	case http.StatusConflict:
		statusText = "Resource conflict."
	case http.StatusBadRequest:
		statusText = "Bad request."
	default:
		statusText = "Error."
	}

	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: getStatusCode(err),
		StatusText:     statusText,
		ErrorText:      err.Error(),
	}
}

func getStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}
	var ierr *domain.Error
	if !errors.As(err, &ierr) {
		return http.StatusInternalServerError
	} else {
		switch ierr.Code() {
		case domain.ErrInternalServerError:
			return http.StatusInternalServerError
		case domain.ErrNotFound:
			return http.StatusNotFound
		case domain.ErrConflict:
			return http.StatusConflict
		case domain.ErrBadParamInput:
			return http.StatusBadRequest
		default:
			return http.StatusInternalServerError
		}
	}

}

func translateError(err error, trans ut.Translator) (errs []error) {
	if err == nil {
		return nil
	}
	validatorErrs := err.(validator.ValidationErrors)
	for _, e := range validatorErrs {
		translatedErr := fmt.Errorf(e.Translate(trans))
		errs = append(errs, translatedErr)
	}
	return errs
}
