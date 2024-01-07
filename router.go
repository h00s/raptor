package raptor

type routes []route

type route struct {
	Method     string
	Path       string
	Controller string
	Action     string
}

type Router struct {
	routes routes
}

func Route(method, path, controller, action string) route {
	return route{
		Method:     method,
		Path:       path,
		Controller: controller,
		Action:     action,
	}
}

func Routes(r ...route) *Router {
	router := &Router{
		routes: routes{},
	}
	for _, route := range r {
		router.routes = append(router.routes, route)
	}
	return router
}
