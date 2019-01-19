package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func errorResponse(w http.ResponseWriter, msg string, code int) {

	w.Header().Set("Content-type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Method", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("AAccess-Control-Request-Headers", "*")

	response := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{"err", msg}

	e, _ := json.Marshal(response)
	w.WriteHeader(code)
	w.Write(e)
	return
}

func successResponse(w http.ResponseWriter, data interface{}, code int) {

	w.Header().Set("Content-type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("AAccess-Control-Request-Headers", "*")
	w.Header().Set("Access-Control-Allow-Method", "*")

	response := struct {
		Status  string      `json:"status"`
		Details interface{} `json:"details"`
	}{"ok", data}

	o, _ := json.Marshal(response)
	w.Write(o)
	return
}

func router(port int) {

	r := mux.NewRouter()

	routes := []struct {
		route   string
		handler func(http.ResponseWriter, *http.Request)
		method  string
	}{
		{
			route:  "/api/v1/{year}/summary",
			method: "GET",
			handler: func(w http.ResponseWriter, r *http.Request) {

				year := mux.Vars(r)["year"]

				summary := school(year)

				successResponse(w, summary, 200)

			},
		},
		{
			route:  "/api/v1/events",
			method: "GET",
			handler: func(w http.ResponseWriter, r *http.Request) {

				events := getEvents()

				successResponse(w, events, 200)
			},
		},
		{
			route:  "/api/v1/{year}/terms",
			method: "GET",
			handler: func(w http.ResponseWriter, r *http.Request) {
				year := mux.Vars(r)["year"]

				successResponse(w, getTerms(year), 200)

			},
		}, {
			route:  "/",
			method: "GET",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.FileServer(http.Dir("static"))
			},
		}, {
			route:  "/*",
			method: "OPTIONS",
			handler: func(w http.ResponseWriter, r *http.Request) {
				successResponse(w, map[string]string{}, 200)
			},
		},
	}

	for _, p := range routes {
		fmt.Printf("    - Adding handler for '%s %s'\n", p.method, p.route)
		r.HandleFunc(p.route, p.handler).Methods(p.method)
	}

	r.PathPrefix("/api").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorResponse(w, "Not Found", 404)
	})

	http.ListenAndServe(fmt.Sprintf(":%d", _port), r)
}
