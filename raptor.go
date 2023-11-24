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
	config      Config
	server      *fiber.App
	controllers map[string]*Controller
	routes      Routes
	Services    *Services
}

func NewMVCRaptor(userConfig ...Config) *Raptor {
	server := newMVCServer()

	raptor := &Raptor{
		config:      config(userConfig...),
		server:      server,
		controllers: make(map[string]*Controller),
		Services:    NewServices(),
		routes:      nil,
	}

	return raptor
}

func NewAPIRaptor(userConfig ...Config) *Raptor {
	server := newAPIServer()

	raptor := &Raptor{
		config:   config(userConfig...),
		server:   server,
		Services: NewServices(),
		routes:   nil,
	}

	return raptor
}

func (r *Raptor) Start() {
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

func newMVCServer() *fiber.App {
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

func newAPIServer() *fiber.App {
	server := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

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

/*func (r *Raptor) registerRoutes() {
	for _, controllerRoute := range r.Router.ControllerRoutes {
		r.registerController(controllerRoute.Controller)

		for _, route := range controllerRoute.Routes {
			r.server.Add(route.Method, route.Path, wrapHandler(route.Handler))
		}
	}
}*/

func (r *Raptor) RegisterControllers(controllers Controllers) {
	r.controllers = controllers
	for _, controller := range r.controllers {
		controller.SetServices(r)
	}
}

func (r *Raptor) RegisterRoutes(routes Routes) {
	r.routes = routes
	for _, route := range r.routes {
		r.Route(route.Method, route.Path, route.Controller, route.Action)
	}
}

func (r *Raptor) Route(method, path, controller, action string) {
	r.server.Add(method, path, wrapHandler(r.controllers[controller].Actions[action].Function))
}
