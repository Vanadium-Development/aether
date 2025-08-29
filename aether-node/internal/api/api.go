package api

import (
	"fmt"
	"log"
	"net/http"
	"node/internal/state"
	"strconv"
)

type State struct {
	node *state.Node
}

func (state *State) getRoot(writer http.ResponseWriter, req *http.Request) {
	log.Println("GET /")
	_, _ = fmt.Fprintf(writer, "Aether node %s is up and running!", state.node.ID)
}

func InitializeApi(port uint16, node state.Node) {
	s := &State{&node}
	http.HandleFunc("/", s.getRoot)

	log.Printf("Listening on port %d\n", port)
	err := http.ListenAndServe(":"+strconv.Itoa(int(port)), nil)
	if err != nil {
		log.Printf("Error initializing API: %s\n", err)
		return
	}
}
