// Package server implements http server lifecycle and all available APIs
package server

import (
	"agentsmith/src/logger"
	"context"
	"embed"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

var log = logger.Logger("server", 1, 1, 1)

func StartServer(fsEmbed embed.FS, port string, readyCh chan string) {
	log.D("Starting agent server")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	if logger.DEBUG == 1 {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(CORSMiddleware)

	uiFS, err := fs.Sub(fsEmbed, "src/ui")
	log.CheckE(err, nil, "Failed to create sub FS for ui")
	router.StaticFS("/ui", http.FS(uiFS))

	if logger.DEBUG == 1 {
		router.Use(gin.Logger())
	}

	server := &http.Server{
		Handler:      router,
		ReadTimeout:  0,
		WriteTimeout: 0,
		IdleTimeout:  0,
	}
	listener, err := net.Listen("tcp", "127.0.0.1:"+port)
	log.CheckE(err, nil, "Failed to bind server port")
	addr := listener.Addr().String()

	initRoutes(router, server)

	go func() {
		err := server.Serve(listener)
		log.CheckE(err, nil, "Server listen failed")
	}()

	log.D("Server started at ", addr)
	readyCh <- addr

	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()
	server.Shutdown(ctx)
}

func initRoutes(router *gin.Engine, server *http.Server) {
	InitAgentRoutes(router)

	if logger.DEBUG == 1 {
		InitDebugRoutes(router, server)
	}
}

func CORSMiddleware(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
	c.Header("Access-Control-Allow-Headers", "authorization, origin, content-type, accept")
	c.Header("Allow", "HEAD,GET,POST,PUT,PATCH,DELETE,OPTIONS")
	if c.Request.Method != "OPTIONS" {
		c.Next()
	} else {
		c.AbortWithStatus(http.StatusOK)
	}
}
