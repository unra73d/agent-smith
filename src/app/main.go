package main

import (
	"agentsmith/src/agent"
	"agentsmith/src/logger"
	"agentsmith/src/server"
	"embed"
	"fmt"
	"net"
	"net/http"
	"os"

	webview "github.com/webview/webview_go"
)

var log = logger.Logger("main", 1, 1, 1)

//go:embed ui/*
var uiFS embed.FS

func main() {
	defer logger.BreakOnError()

	os.Setenv("AS_BACKEND_PORT", "8008")
	os.Setenv("AS_AGENT_DB_FILE", "app.db")

	agent.LoadAgent()

	// agent api server
	go server.StartServer()

	// server that serves UI
	mux := http.NewServeMux()

	// Static File Server for UI assets
	// Determine the directory of the executable or use current dir during dev
	// ex, err := os.Executable()
	// if err != nil {
	// 	panic(err)
	// }
	// exePath := filepath.Dir(ex)
	// Use relative path for development, switch to exePath for bundled app
	// uiDir := "./ui" // Or: filepath.Join(exePath, "ui") for bundled app
	//log.D("Serving static files from: ", uiFS)
	fs := http.FileServer(http.FS(uiFS))
	// Serve everything under / that isn't an API route
	mux.Handle("/", fs)

	// --- 2. Start HTTP Server on a Free Port ---
	// Listen on a random available port on localhost. This avoids conflicts.
	listener, err := net.Listen("tcp", "127.0.0.1:0") // ":0" means random port
	log.CheckE(err, nil, "Failed to bind UI port")

	serverAddr := listener.Addr().String()
	serverURL := fmt.Sprintf("http://%s/ui/", serverAddr)
	log.D("UI server starting on: ", serverAddr)

	// Start the server in a goroutine so it doesn't block the WebView
	go func() {
		err := http.Serve(listener, mux)
		log.CheckE(err, nil, "Failed to start UI server")
	}()

	// --- 3. Setup and Run WebView ---
	w := webview.New(logger.DEBUG == 1)
	defer w.Destroy() // Ensure cleanup

	w.SetTitle("Agent Smith")
	w.SetSize(1000, 800, webview.HintNone)

	// Navigate the webview to the local server's URL
	log.D("Navigating WebView to: ", serverURL)
	w.Navigate(serverURL)

	// Run the WebView event loop (this blocks until the window is closed)
	w.Run()

	log.D("App closed")
}
