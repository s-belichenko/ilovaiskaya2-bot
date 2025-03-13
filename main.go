package main

import (
	"fmt"
	"io"
	"net/http"
)

// Handler обработчик для корректной работы serverless-функции
func Handler(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("X-Custom-Header", "Test")
	rw.WriteHeader(200)
	name := req.URL.Query().Get("name")
	io.WriteString(rw, fmt.Sprintf("Hello, %s!", name))
}
