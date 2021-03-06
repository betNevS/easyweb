package core

import (
	"log"
	"net/http"
	"strings"
)

type Core struct {
	router      map[string]*Tree
	middlewares []ControllerHandler
}

func New() *Core {
	router := map[string]*Tree{}
	router["GET"] = NewTree()
	router["POST"] = NewTree()
	router["PUT"] = NewTree()
	router["DELETE"] = NewTree()

	return &Core{
		router: router,
	}
}

func (c *Core) Use(middleware ...ControllerHandler) {
	c.middlewares = append(c.middlewares, middleware...)
}

func (c *Core) Get(uri string, handlers ...ControllerHandler) {
	allHandlers := append(c.middlewares, handlers...)
	if err := c.router["GET"].AddRouter(uri, allHandlers); err != nil {
		log.Fatal(err)
	}
}

func (c *Core) Post(uri string, handlers ...ControllerHandler) {
	allHandlers := append(c.middlewares, handlers...)
	if err := c.router["POST"].AddRouter(uri, allHandlers); err != nil {
		log.Fatal(err)
	}
}

func (c *Core) Put(uri string, handlers ...ControllerHandler) {
	allHandlers := append(c.middlewares, handlers...)
	if err := c.router["PUT"].AddRouter(uri, allHandlers); err != nil {
		log.Fatal(err)
	}
}

func (c *Core) Delete(uri string, handlers ...ControllerHandler) {
	allHandlers := append(c.middlewares, handlers...)
	if err := c.router["DELETE"].AddRouter(uri, allHandlers); err != nil {
		log.Fatal(err)
	}
}

func (c *Core) FindRouteByRequest(request *http.Request) []ControllerHandler {
	uri := request.URL.Path
	method := request.Method
	upperMethod := strings.ToUpper(method)

	if methodHandlers, ok := c.router[upperMethod]; ok {
		return methodHandlers.FindHandler(uri)
	}

	return nil
}

func (c *Core) FindRouteNodeByRequest(request *http.Request) *node {
	uri := request.URL.Path
	method := request.Method
	upperMethod := strings.ToUpper(method)
	if methodHandlers, ok := c.router[upperMethod]; ok {
		return methodHandlers.FindNode(uri)
	}
	return nil
}

func (c *Core) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	ctx := NewContext(request, response)

	node := c.FindRouteNodeByRequest(request)
	if node == nil {
		ctx.SetStatus(http.StatusNotFound).JSON("NOT FOUND ROUTER")
		return
	}

	ctx.SetHandlers(node.handlers)

	params := node.parseParamsFromEndNode(request.URL.Path)
	ctx.SetParams(params)

	if err := ctx.Next(); err != nil {
		ctx.SetStatus(http.StatusInternalServerError).JSON("INNER ERROR")
		return
	}
}

func (c *Core) Group(prefix string) IGroup {
	return NewGroup(c, prefix)
}
