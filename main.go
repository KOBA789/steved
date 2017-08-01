package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/koba789/steved/task"
	"log"
	"net/http"
	"os"
)

func jobHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	task, ok, err := task.GetTask(name)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	envMap := make(map[string]string)
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err = decoder.Decode(&envMap)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	env := []string{}
	for key, value := range envMap {
		env = append(env, key+"="+value)
	}

	err = task.Spawn(name, env)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/jobs/{name}", jobHandler).Methods("POST")
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8000"
	}
	host, ok := os.LookupEnv("HOST")
	if !ok {
		host = ""
	}
	log.Fatal(http.ListenAndServe(host+":"+port, r))
}
