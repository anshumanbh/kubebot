package main

import "net/http"

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		Index,
	},
	Route{
		"API v1 POST Test",
		"POST",
		"/api/v1",
		Apitest,
	},
	Route{
		"API v1 GET Test",
		"GET",
		"/api/v1",
		Apitest,
	},
	Route{
		"RunTool",
		"POST",
		"/api/v1/runtool",
		RunTool,
	},
	Route{
		"RunAutomation",
		"POST",
		"/api/v1/runautomation",
		RunAutomation,
	},
}
