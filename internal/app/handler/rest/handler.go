package rest

import (
	"fmt"
	"hash/crc32"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/karlseguin/ccache"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sku4/ad-api/internal/app"
	"github.com/sku4/ad-api/internal/app/handler/rest/middleware"
	"github.com/sku4/ad-api/internal/app/service"
)

const (
	ttlCacheSize = 100
)

type Handler struct {
	services service.Service
	cache    *ccache.Cache
	crcTable *crc32.Table
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{
		services: *services,
		cache:    ccache.New(ccache.Configure().MaxSize(ttlCacheSize)),
		crcTable: crc32.MakeTable(crc32.IEEE),
	}
}

func (h *Handler) InitRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(chiMiddleware.RequestID)

	r.Handle("/metrics", promhttp.Handler())
	r.Group(func(r chi.Router) {
		r.Use(chiMiddleware.Logger)
		r.Use(middleware.Metrics)

		r.Post("/ads", h.Ads)
		r.Post("/locations", h.Locations)
		r.Get("/streets", h.Streets)

		r.Group(func(r chi.Router) {
			r.Get("/", h.Map)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.ValidHash)
			r.Get(fmt.Sprintf("/%s", app.AddSub), h.AddSub)
			r.Post(fmt.Sprintf("/%s/subscription/add", app.AddSub), h.SubscriptionAdd)
		})
	})

	fileServer(r, "/static", http.Dir("./web/static"))

	return r
}

// fileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func fileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		ctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(ctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(vfs{root}))
		fs.ServeHTTP(w, r)
	})
}

type vfs struct {
	fs http.FileSystem
}

func (v vfs) Open(path string) (http.File, error) {
	f, err := v.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, errOpen := v.fs.Open(index); errOpen != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}

			return nil, errOpen
		}
	}

	return f, nil
}
