package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/tesrohit-developer/rest-server-scrathpad/plugin"
	"log"
	"net/http"
)

var pluginInterface interface{}

func getSidelinePlugin() interface{} {
	s := plugin.NewManager("sideline_plugin", "sideline-*", "./plugins/built", &plugin.CheckMessageSidelineImplPlugin{})
	//defer s.Dispose()

	err := s.Init()

	if err != nil {
		log.Fatal(err.Error())
	}

	s.Launch()

	p, err := s.GetInterface("em")
	if err != nil {
		log.Fatal(err.Error())
	}
	return p
}

func echoString(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		fmt.Println("id is missing in parameters")
	}
	fmt.Println(`id := `, id)
	/*s1 := "em1"
	s1bytes, _ := json.Marshal(s1)
	pluginInterface.(plugin.CheckMessageSidelineImpl).CheckMessageSideline(s1bytes)*/
	fmt.Fprintf(w, "hello"+id)
}

func main() {
	fmt.Println("Starting server")
	r := mux.NewRouter()
	//pluginInterface = getSidelinePlugin()
	r.HandleFunc("/{id}", echoString)
	log.Fatal(http.ListenAndServe(":8081", r))
}
