package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Display Human-Readable information about the node
func (ctx *RouteCtx) getRootHandler(writer http.ResponseWriter, req *http.Request) {
	_, _ = fmt.Fprintf(writer, "%s\n\n--------------------\nAether node is up and running!\nPort: %d\nName: %s\nUUID: %s\n--------------------", banner, ctx.Port, ctx.Node.Name, ctx.Node.ID)
}

func (ctx *RouteCtx) getInfoHandler(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(writer).Encode(ctx.Node.NodeInfoMap())
	_, _ = fmt.Fprintf(writer, "")
}
