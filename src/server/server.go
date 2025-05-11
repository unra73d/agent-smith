// Package server implements http server lifecycle and all available APIs
package server

import (
	"agentsmith/src/logger"
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

var log = logger.Logger("server", 1, 1, 1)

const listenPort = 8008

func StartServer() {
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

	if logger.DEBUG == 1 {
		router.Use(gin.Logger())
	}

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(listenPort),
		Handler: router,
	}

	initRoutes(router, server)

	go func() {
		err := server.ListenAndServe()
		log.CheckE(err, nil, "Server listen failed")
	}()

	log.D("Server started")

	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
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
	c.Header("Content-Type", "application/json")
	if c.Request.Method != "OPTIONS" {
		c.Next()
	} else {
		c.AbortWithStatus(http.StatusOK)
	}
}
