package api

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

func (ctx *RouteCtx) handleFallbackInfoPage(writer http.ResponseWriter, request *http.Request) {
	_, _ = fmt.Fprintf(writer, "%s\n\n--------------------\nAether node is up and running!\nPort: %d\nName: %s\nUUID: %s\n--------------------", banner, ctx.Port, ctx.Node.Name, ctx.Node.ID)
}

// Return information about current node as a human-readable page
func (ctx *RouteCtx) getRootHandler(writer http.ResponseWriter, req *http.Request) {
	tmpl, err := template.ParseFiles("static/index.html")

	if err != nil {
		goto fallback
	}

	err = tmpl.Execute(writer, map[string]interface{}{
		"UUID":      ctx.Node.ID.String(),
		"Name":      ctx.Node.Name,
		"Port":      strconv.Itoa(int(ctx.Port)),
		"NodeColor": template.CSS(fmt.Sprintf("rgb(%d,%d,%d)", ctx.Node.Color.R, ctx.Node.Color.G, ctx.Node.Color.B)),
	})

	if err != nil {
		goto fallback
	}

	return

fallback:
	ctx.handleFallbackInfoPage(writer, req)
	log.Printf("Could not parse template: %s\n", err)
	return
}

// Return information about current node as JSON
func (ctx *RouteCtx) getInfoHandler(writer http.ResponseWriter, req *http.Request) {
	RespondJson(writer, ctx.Node.NodeInfoMap())
}
