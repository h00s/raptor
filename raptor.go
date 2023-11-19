package raptor

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/jet/v2"
)

type Raptor struct {
	config   Config
	server   *fiber.App
	Services *Services
	Router   *Router
}

func NewRaptor(config ...Config) *Raptor {
	server := newServer()

	raptor := &Raptor{
		config:   Config{},
		server:   server,
		Services: NewServices(),
		Router:   NewRouter(),
	}

	if len(config) > 0 {
		raptor.config = config[0]
	}

	if raptor.config.Address == "" {
		raptor.config.Address = DefaultAddress
	}
	if raptor.config.Port == 0 {
		raptor.config.Port = DefaultPort
	}

	return raptor
}

func (r *Raptor) Start() {
	r.RegisterRoutes(r.Router)
	r.Services.Log.Info("====> Starting Raptor <====")
	if r.checkPort() {
		go func() {
			if err := r.server.Listen(r.address()); err != nil && err != http.ErrServerClosed {
				panic(err)
			}
		}()
		r.info()
		r.waitForShutdown()
	} else {
		r.Services.Log.Error(fmt.Sprintf("Unable to bind on address %s, already in use!", r.address()))
	}
}

func (r *Raptor) address() string {
	return r.config.Address + ":" + fmt.Sprint(r.config.Port)
}

func (r *Raptor) checkPort() bool {
	ln, err := net.Listen("tcp", r.address())
	if err == nil {
		ln.Close()
	}
	return err == nil
}

func newServer() *fiber.App {
	engine := jet.New("app/views", ".html.jet")

	// TODO: add this to the config
	engine.Reload(true)

	server := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		Views:                 engine,
		ViewsLayout:           "layouts/main",
	})

	server.Static("/public", "./public")
	return server
}

func (r *Raptor) info() {
	r.Services.Log.Info("Raptor is running! 🎉")
	r.Services.Log.Info(fmt.Sprintf("Listening on http://%s", r.address()))
}

func (r *Raptor) waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	r.Services.Log.Warn("Shutting down Raptor...")
	if err := r.server.Shutdown(); err != nil {
		r.Services.Log.Error("Server Shutdown:", err)
	}
	r.Services.Log.Warn("Raptor exited, bye bye!")
}

func (r *Raptor) RegisterController(c Controller) {
	c.SetServices(r)
}

func (r *Raptor) RegisterRoutes(router *Router) {
	for _, route := range router.Routes {
		r.server.Add(route.Method, route.Path, wrapHandler(route.Handler))
	}
}
