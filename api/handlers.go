package api

import (
	"context"
	"errors"
	"fmt"
	"lintang/navigatorx/alg"
	"lintang/navigatorx/service"
	"lintang/navigatorx/types"
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
	ManyToManyQuery(ctx context.Context, sourcesLat, sourcesLon, destsLat, destsLon []float64) map[alg.Coordinate][]service.TargetResult

	TravelingSalesmanProblemSimulatedAnneal(ctx context.Context, citiesLat []float64, citiesLon []float64) ([]alg.Coordinate, string, float64, float64)
	WeightedBipartiteMatching(ctx context.Context, riderLatLon map[string][]float64, driverLatLon map[string][]float64) (map[string]map[string]float64, float64, error)
}

type NavigationHandler struct {
	svc          NavigationService
	promeMetrics *metrics
}

func NavigatorRouter(r *chi.Mux, svc NavigationService, m *metrics) {
	handler := &NavigationHandler{svc, m}

	r.Group(func(r chi.Router) {
		r.Route("/api/navigations", func(r chi.Router) {
			r.Post("/shortest-path", handler.shortestPathETA)
			r.Post("/shortest-path-alternative-street", handler.shortestPathAlternativeStreetETA)
			r.Post("/shortest-path-ch", handler.shortestPathETACH)
			r.Post("/map-matching", handler.HiddenMarkovModelMapMatching)
			r.Post("/many-to-many", handler.ManyToManyQuery)
			r.Post("/tsp", handler.TravelingSalesmanProblemSimulatedAnnealing)
			r.Post("/matching", handler.WeightedBipartiteMatching)
		})
	})
}

// SortestPathRequest model info
//
//	@Description	request body untuk shortest path query antara 2 tempat di openstreetmap
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

// SortestPathAlternativeStreetRequest model info
//
//	@Description	request body untuk shortest path query antara banyak source dan banyak destination di openstreetmap
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

// ShortestPathResponse	model info
//
//	@Description	response body untuk shortest path query antara 2 tempat di openstreetmap
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

// shortestPathETA
//
//	@Summary		shortest path query antara 2 tempat di openstreetmap.
//	@Description	shortest path query antara 2 tempat di openstreetmap. Hanya 1 source dan 1 destination
//	@Tags			navigations
//	@Param			body	body	SortestPathRequest	true	"request body query shortest path antara 2 tempat"
//	@Accept			application/json
//	@Produce		application/json
//	@Router			/navigations/shortest-path [post]
//	@Success		200	{object}	ShortestPathResponse
//	@Failure		400	{object}	ErrResponse
//	@Failure		500	{object}	ErrResponse
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

	h.promeMetrics.SPQueryCount.WithLabelValues("true").Inc()
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

// shortestPathAlternativeStreetETA
//
//	@Summary		shortest path query antara 2 tempat di openstreetmap dengan menentukan alternative street untuk rutenya.
//	@Description	shortest path query antara 2 tempat di openstreetmap dengan menentukan alternative street untuk rutenya.. Hanya 1 source dan 1 destination
//	@Tags			navigations
//	@Param			body	body	SortestPathAlternativeStreetRequest	true	"request body query shortest path antara 2 tempat"
//	@Accept			application/json
//	@Produce		application/json
//	@Router			/navigations/shortest-path-alternative-street [post]
//	@Success		200	{object}	ShortestPathResponse
//	@Failure		400	{object}	ErrResponse
//	@Failure		500	{object}	ErrResponse
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

	h.promeMetrics.SPQueryCount.WithLabelValues("true").Inc()
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

// shortestPathETACH
//
//	@Summary		shortest path query antara 2 tempat di openstreetmap  pake bidirectional dijkstra only, sebenarnya di endpoint shortest-path juga pake bidirectional dijkstra. jika preprocessing contraction hierarchiesnya selesai
//	@Description	shortest path query antara 2 tempat di openstreetmap  pake bidirectional dijkstra only, sebenarnya di endpoint shortest-path juga pake bidirectional dijkstra. jika preprocessing contraction hierarchiesnya selesai. Hanya 1 source dan 1 destination
//	@Tags			navigations
//	@Param			body	body	SortestPathRequest	true	"request body query shortest path antara 2 tempat"
//	@Accept			application/json
//	@Produce		application/json
//	@Router			/navigations/shortest-path-ch [post]
//	@Success		200	{object}	ShortestPathResponse
//	@Failure		400	{object}	ErrResponse
//	@Failure		500	{object}	ErrResponse
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

	h.promeMetrics.SPQueryCount.WithLabelValues("true").Inc()
	p, n, route, eta, dist, err := h.svc.ShortestPathETACH(r.Context(), data.SrcLat, data.SrcLon, data.DstLat, data.DstLon)
	if err != nil {

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

// MapMatchingRequest model info
//
//	@Description	request body untuk map matching pakai hidden markov model
type MapMatchingRequest struct {
	Coordinates []Coord `json:"coordinates" validate:"required,dive"`
}

// Coord model info
//
//	@Description	model untuk koordinat
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

// MapMatchingResponse model info
//
//	@Description	response body untuk map matching pakai hidden markov model
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

// HiddenMarkovModelMapMatching
//
//	@Summary		map matching pakai hidden markov model. Snapping noisy GPS coordinates ke road network lokasi asal gps seharusnya
//	@Description	map matching pakai hidden markov model. Snapping noisy GPS coordinates ke road network lokasi asal gps seharusnya
//	@Tags			navigations
//	@Param			body	body	MapMatchingRequest	true	"request body hidden markov model map matching"
//	@Accept			application/json
//	@Produce		application/json
//	@Router			/navigations/map-matching [post]
//	@Success		200	{object}	MapMatchingResponse
//	@Failure		400	{object}	ErrResponse
//	@Failure		500	{object}	ErrResponse
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

// ManyToManyQueryRequest model info
//
//	@Description	response body untuk query shortest path many to many
type ManyToManyQueryRequest struct {
	Sources []Coord `json:"sources" validate:"required,dive"`
	Targets []Coord `json:"targets" validate:"required,dive"`
}

func (s *ManyToManyQueryRequest) Bind(r *http.Request) error {
	if len(s.Sources) == 0 || len(s.Targets) == 0 {
		return errors.New("invalid request")
	}
	return nil
}

// NodeRes model info
//
//	@Description	model untuk node coordinate
type NodeRes struct {
	Lat float64 `json:"lat" `
	Lon float64 `json:"lon" `
}

// TargetRes model info
//
//	@Description	model untuk destinations di query shortest path many to many
type TargetRes struct {
	Target NodeRes `json:"target"`
	Path   string  `json:"path"`
	Dist   float64 `json:"distance"`
	ETA    float64 `json:"ETA"`
}

// SrcTargetPair model info
//
//	@Description	model untuk mapping source dan target di query shortest path many to many
type SrcTargetPair struct {
	Source  NodeRes     `json:"source"`
	Targets []TargetRes `json:"targets"`
}

// ManyToManyQueryResponse model info
//
//	@Description	response body untuk query shortest path many to many
type ManyToManyQueryResponse struct {
	Results []SrcTargetPair `json:"results"`
}

// ManyToManyQuery
//
//	@Summary		many to many query shortest path . punya banyak source dan banyak destination buat querynya. Mencari shortesth path ke setiap destination untuk setiap source
//	@Description	many to many query shortest path . punya banyak source dan banyak destination buat querynya
//	@Tags			navigations
//	@Param			body	body	ManyToManyQueryRequest	true	"request body query shortest path many to many"
//	@Accept			application/json
//	@Produce		application/json
//	@Router			/navigations/many-to-many [post]
//	@Success		200	{object}	ManyToManyQueryResponse
//	@Failure		400	{object}	ErrResponse
//	@Failure		500	{object}	ErrResponse
func (h *NavigationHandler) ManyToManyQuery(w http.ResponseWriter, r *http.Request) {
	data := &ManyToManyQueryRequest{}
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

	sourcesLat, sourcesLon, destsLat, destsLon := []float64{}, []float64{}, []float64{}, []float64{}
	for _, s := range data.Sources {
		sourcesLat = append(sourcesLat, s.Lat)
		sourcesLon = append(sourcesLon, s.Lon)
	}
	for _, d := range data.Targets {
		destsLat = append(destsLat, d.Lat)
		destsLon = append(destsLon, d.Lon)
	}

	results := h.svc.ManyToManyQuery(r.Context(), sourcesLat, sourcesLon, destsLat, destsLon)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, RenderManyToManyQueryResponse(results))
}

func RenderManyToManyQueryResponse(res map[alg.Coordinate][]service.TargetResult) *ManyToManyQueryResponse {
	results := []SrcTargetPair{}
	for k, v := range res {

		targets := []TargetRes{}
		for _, t := range v {
			targets = append(targets, TargetRes{
				Target: NodeRes{
					Lat: t.TargetCoord.Lat,
					Lon: t.TargetCoord.Lon,
				},
				Path: t.Path,
				Dist: t.Dist,
				ETA:  t.ETA,
			},
			)
		}
		results = append(results, SrcTargetPair{
			Source: NodeRes{
				Lat: k.Lat,
				Lon: k.Lon,
			},
			Targets: targets,
		})
	}
	return &ManyToManyQueryResponse{
		Results: results,
	}
}

// TravelingSalesmanProblemRequest model info
//
//	@Description	request body untuk traveling salesman problem query
type TravelingSalesmanProblemRequest struct {
	CitiesCoord []Coord `json:"cities_coord" validate:"required,dive"`
}

func (s *TravelingSalesmanProblemRequest) Bind(r *http.Request) error {
	if len(s.CitiesCoord) < 2 {
		return errors.New("invalid request")
	}
	return nil
}

// TravelingSalesmanProblemResponse model info
//
//	@Description	response body untuk traveling salesman problem query
type TravelingSalesmanProblemResponse struct {
	Path   string           `json:"path"`
	Dist   float64          `json:"distance"`
	ETA    float64          `json:"ETA"`
	Cities []alg.Coordinate `json:"cities_order"`
}

// TravelingSalesmanProblemSimulatedAnnealing
//
//	@Summary		query traveling salesman problem pakai simulated annealing. Shortest path untuk rute mengunjungi beberapa tempat tepat sekali dan kembali ke tempat asal
//	@Description	query traveling salesman problem pakai simulated annealing. Shortest path untuk rute mengunjungi beberapa tempat tepat sekali dan kembali ke tempat asal
//	@Tags			navigations
//	@Param			body	body	TravelingSalesmanProblemRequest	true	"request body query tsp"
//	@Accept			application/json
//	@Produce		application/json
//	@Router			/navigations/tsp [post]
//	@Success		200	{object}	TravelingSalesmanProblemResponse
//	@Failure		400	{object}	ErrResponse
//	@Failure		500	{object}	ErrResponse
func (h *NavigationHandler) TravelingSalesmanProblemSimulatedAnnealing(w http.ResponseWriter, r *http.Request) {
	data := &TravelingSalesmanProblemRequest{}
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

	citiesLat, citiesLon := []float64{}, []float64{}
	for _, c := range data.CitiesCoord {
		citiesLat = append(citiesLat, c.Lat)
		citiesLon = append(citiesLon, c.Lon)
	}

	tspTourNodes, path, dist, eta := h.svc.TravelingSalesmanProblemSimulatedAnneal(r.Context(), citiesLat, citiesLon)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, RenderTravelingSalesmanProblemResponse(path, dist, eta, tspTourNodes))
}

func RenderTravelingSalesmanProblemResponse(path string, dist float64, eta float64, cities []alg.Coordinate) *TravelingSalesmanProblemResponse {
	return &TravelingSalesmanProblemResponse{
		Path:   path,
		Dist:   dist,
		ETA:    eta,
		Cities: cities,
	}
}

// WeightedBipartiteMatching
// WeightedBipartiteMatchingRequest model info
//
//	@Description	request body untuk rider driver matching (weighted bipartite matching) query
type WeightedBipartiteMatchingRequest struct {
	RiderLatLon  map[string]Coord `json:"rider_lat_lon" validate:"required,dive"`
	DriverLatLon map[string]Coord `json:"driver_lat_lon" validate:"required,dive"`
}

func (s *WeightedBipartiteMatchingRequest) Bind(r *http.Request) error {
	if len(s.DriverLatLon) < 1 || len(s.RiderLatLon) < 1 {
		return errors.New("invalid request")
	}
	return nil
}

// WeightedBipartiteMatchingResponse model info
//
//	@Description	response body untuk rider driver matching query
type WeightedBipartiteMatchingResponse struct {
	Match    map[string]map[string]float64 `json:"match"`
	TotalETA float64                       `json:"total_eta"`
}

// WeightedBipartiteMatching
//
//	@Summary		query weighted bipartite matching. Misalnya, untuk assign beberapa rider ke driver di suatu area secara optimal (untuk backend aplikasi ride hailing).
//	@Description	query weighted bipartite matching. Misalnya, untuk assign beberapa rider ke driver di suatu area secara optimal (untuk backend aplikasi ride hailing).
//	@Tags			navigations
//	@Param			body	body	WeightedBipartiteMatchingRequest	true	"request body query weighted bipartite matching"
//	@Accept			application/json
//	@Produce		application/json
//	@Router			/navigations/matching [post]
//	@Success		200	{object}	WeightedBipartiteMatchingResponse
//	@Failure		400	{object}	ErrResponse
//	@Failure		500	{object}	ErrResponse
func (h *NavigationHandler) WeightedBipartiteMatching(w http.ResponseWriter, r *http.Request) {
	data := &WeightedBipartiteMatchingRequest{}
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

	riderLatLon, driverLatLon := map[string][]float64{}, map[string][]float64{}
	for k, v := range data.RiderLatLon {
		riderLatLon[k] = []float64{v.Lat, v.Lon}
	}
	for k, v := range data.DriverLatLon {
		driverLatLon[k] = []float64{v.Lat, v.Lon}
	}

	match, totalEta, err := h.svc.WeightedBipartiteMatching(r.Context(), riderLatLon, driverLatLon)
	if err != nil {
		render.Render(w, r, ErrChi(err))
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, RenderWeightedBipartiteMatchingResponse(match, totalEta))
}

func RenderWeightedBipartiteMatchingResponse(match map[string]map[string]float64 , totalEta float64) *WeightedBipartiteMatchingResponse {
	return &WeightedBipartiteMatchingResponse{
		Match: match,
		TotalETA: totalEta,
	}
}

// ErrResponse model info
//
//	@Description	model untuk error response
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
	var ierr *types.Error
	if !errors.As(err, &ierr) {
		return http.StatusInternalServerError
	} else {
		switch ierr.Code() {
		case types.ErrInternalServerError:
			return http.StatusInternalServerError
		case types.ErrNotFound:
			return http.StatusNotFound
		case types.ErrConflict:
			return http.StatusConflict
		case types.ErrBadParamInput:
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
