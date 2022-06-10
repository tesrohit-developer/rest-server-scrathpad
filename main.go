package main

import (
	"encoding/json"
	"fmt"
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
	s1 := "em1"
	s1bytes, _ := json.Marshal(s1)
	pluginInterface.(plugin.CheckMessageSidelineImpl).CheckMessageSideline(s1bytes)
	fmt.Fprintf(w, "hello")
}

func main() {
	pluginInterface = getSidelinePlugin()
	http.HandleFunc("/", echoString)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
