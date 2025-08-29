package api

import (
	"fmt"
	"log"
	"net/http"
	"node/internal/state"
	"strconv"
)

const banner = `
  ___       _   _               
 / _ \     | | | |              
/ /_\ \ ___| |_| |__   ___ _ __ 
|  _  |/ _ \ __| '_ \ / _ \ '__|
| | | |  __/ |_| | | |  __/ |   
\_| |_/\___|\__|_| |_|\___|_|   
                                `

type RouteCtx struct {
	Port uint16
	Node *state.Node
}

func registerApiRoutes(state *RouteCtx) {
	http.HandleFunc("/", state.getRootHandler)
	http.HandleFunc("/info", state.getInfoHandler)
}

func InitializeApi(port uint16, node state.Node) {
	s := &RouteCtx{port, &node}
	fmt.Println(banner)
	log.Printf("Aether node is listening on port http://localhost:%d\n", port)

	registerApiRoutes(s)
	err := http.ListenAndServe(":"+strconv.Itoa(int(port)), nil)
	if err != nil {
		log.Printf("Error initializing API: %s\n", err)
		return
	}
}
