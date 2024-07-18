package ngebut

import (
	"errors"
	"github.com/panjf2000/gnet/v2"
	"net/http"
)

type App struct {
	config *Config
	router *Router
}

var DefaultConfig = Config{
	Addr:      ":3000",
	MultiCore: false,
	ErrorHandler: func(c *Context, err error) error {
		status := 500
		message := http.StatusText(status)
		var e *Error
		if errors.As(err, &e) {
			status = e.Status
			message = e.Message
		}

		return c.Status(status).SendString(message)
	},
}

func New(config ...Config) *App {
	c := &DefaultConfig

	if len(config) > 0 {
		if config[0].Addr != "" {
			c.Addr = config[0].Addr
		}

		if config[0].ErrorHandler != nil {
			c.ErrorHandler = config[0].ErrorHandler
		}
	}

	return &App{
		config: c,
		router: &Router{
			routes:   []routeKey{},
			handlers: make(map[routeKey][]HandlerFunc),
		},
	}
}

func (a *App) Router() *Router {
	return a.router
}

func (a *App) Get(path string, handler ...HandlerFunc) {
	a.router.Add(http.MethodGet, path, handler...)
}

func (a *App) Post(path string, handler ...HandlerFunc) {
	a.router.Add(http.MethodPost, path, handler...)
}

func (a *App) Put(path string, handler ...HandlerFunc) {
	a.router.Add(http.MethodPut, path, handler...)
}

func (a *App) Patch(path string, handler ...HandlerFunc) {
	a.router.Add(http.MethodPatch, path, handler...)
}

func (a *App) Delete(path string, handler ...HandlerFunc) {
	a.router.Add(http.MethodDelete, path, handler...)
}

func (a *App) Options(path string, handler ...HandlerFunc) {
	a.router.Add(http.MethodOptions, path, handler...)
}

func (a *App) Head(path string, handler ...HandlerFunc) {
	a.router.Add(http.MethodHead, path, handler...)
}

func (a *App) Run() error {
	hs := &httpServer{
		addr:         a.config.Addr,
		multicore:    a.config.MultiCore,
		router:       a.router,
		errorHandler: a.config.ErrorHandler,
	}

	return gnet.Run(
		hs,
		a.config.Addr,
		gnet.WithLogger(&logger{}),
		gnet.WithMulticore(a.config.MultiCore),
		gnet.WithReusePort(true),
		gnet.WithReuseAddr(true),
	)

}
