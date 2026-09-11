package main

import (
	"flag"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	port := flag.Int("port", 8080, "HTTP server port")
	dir := flag.String("dir", ".", "Directory to serve")
	flag.Parse()

	// Ensure .wasm is mapped to application/wasm
	_ = mime.AddExtensionType(".wasm", "application/wasm")
	_ = mime.AddExtensionType(".js", "application/javascript")
	_ = mime.AddExtensionType(".html", "text/html")
	_ = mime.AddExtensionType(".css", "text/css")

	serveDir, err := filepath.Abs(*dir)
	if err != nil {
		serveDir = *dir
	}

	// Verify gravitysim.wasm exists
	wasmPath := filepath.Join(serveDir, "gravitysim.wasm")
	if _, err := os.Stat(wasmPath); os.IsNotExist(err) {
		// If running from repo root or web/ directory
		if _, err := os.Stat("web/gravitysim.wasm"); err == nil {
			serveDir, _ = filepath.Abs("web")
		}
	}

	fs := http.FileServer(http.Dir(serveDir))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS and cache control headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Cache-Control", "no-cache")
		fs.ServeHTTP(w, r)
	})

	addr := fmt.Sprintf(":%d", *port)
	fmt.Println("================================================================")
	fmt.Println("🪐 Gravity Mass Simulator 3D - WebAssembly Server")
	fmt.Printf("📂 Serving files from: %s\n", serveDir)
	fmt.Printf("🌐 Open in your browser: http://localhost:%d\n", *port)
	fmt.Println("================================================================")

	if err := http.ListenAndServe(addr, handler); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}
