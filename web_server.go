package main

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

func runWebServer(port int) {
	_ = mime.AddExtensionType(".wasm", "application/wasm")
	_ = mime.AddExtensionType(".js", "application/javascript")
	_ = mime.AddExtensionType(".html", "text/html")
	_ = mime.AddExtensionType(".css", "text/css")

	serveDir := "web"
	if _, err := os.Stat("web/gravitysim.wasm"); os.IsNotExist(err) {
		if _, err := os.Stat("gravitysim.wasm"); err == nil {
			serveDir = "."
		}
	}

	absDir, _ := filepath.Abs(serveDir)
	fs := http.FileServer(http.Dir(serveDir))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Cache-Control", "no-cache")
		fs.ServeHTTP(w, r)
	})

	fmt.Println("================================================================")
	fmt.Println("🪐 Gravity Mass Simulator 3D - WebAssembly Server")
	fmt.Printf("📂 Serving directory: %s\n", absDir)
	fmt.Printf("🌐 Open in browser:  http://localhost:%d\n", port)
	fmt.Println("================================================================")

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), handler); err != nil {
		fmt.Printf("Web server error: %v\n", err)
	}
}
