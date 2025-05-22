package main

import (
	"agentsmith/src/agent"
	"agentsmith/src/logger"
	"agentsmith/src/server"
	"embed"
	"flag"
	"fmt"
	"os"

	webview "github.com/webview/webview_go"
)

var log = logger.Logger("main", 1, 1, 1)

//go:embed ui
//go:embed ui/**/*
var uiFS embed.FS

func main() {
	defer logger.BreakOnError()

	serverOnly := flag.Bool("server", false, "Run only the server without UI")
	port := flag.Int("port", 0, "Specify the port for the server to listen on")
	flag.Parse()

	entries, _ := uiFS.ReadDir("ui")
	for _, e := range entries {
		fmt.Println(e.Name())
	}

	os.Setenv("AS_AGENT_DB_FILE", "app.db")

	agent.LoadAgent()

	// agent api server
	serverReadyCh := make(chan string)
	go server.StartServer(uiFS, fmt.Sprintf("%d", *port), serverReadyCh)

	addr := <-serverReadyCh
	serverURL := fmt.Sprintf("http://%s/ui/", addr)

	if *serverOnly {
		log.D("Server running at: ", serverURL)
		select {}
	} else {
		w := webview.New(logger.DEBUG == 1)
		defer w.Destroy() // Ensure cleanup

		w.SetTitle("Agent Smith")
		w.SetSize(1200, 800, webview.HintNone)

		log.D("Navigating WebView to: ", serverURL)
		w.Navigate(serverURL)

		w.Run()
		log.D("App closed")
	}
}
