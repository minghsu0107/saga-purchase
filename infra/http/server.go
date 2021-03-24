package http

import (
	"context"
	"io"
	"net/http"

	watermillHTTP "github.com/ThreeDotsLabs/watermill-http/pkg/http"
	"github.com/gin-gonic/gin"
	conf "github.com/minghsu0107/saga-purchase/config"
	"github.com/minghsu0107/saga-purchase/infra/http/middleware"
	log "github.com/sirupsen/logrus"
)

// Server is the http wrapper
type Server struct {
	Config         *conf.Config
	Engine         *gin.Engine
	Router         *Router
	Svr            *http.Server
	sseRouter      *watermillHTTP.SSERouter
	jwtAuthChecker *middleware.JWTAuthChecker
}

// NewEngine is a factory for gin engine instance
// Global Middlewares and api log configurations are registered here
func NewEngine(config *conf.Config) *gin.Engine {
	gin.SetMode(config.GinMode)
	if config.GinMode == "release" {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.DebugLevel)
	}
	gin.DefaultWriter = io.Writer(config.Logger.Writer)
	log.SetOutput(gin.DefaultWriter)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.LogMiddleware())
	engine.Use(middleware.CORSMiddleware())
	return engine
}

// NewServer is the factory for server instance
func NewServer(config *conf.Config, engine *gin.Engine, router *Router, sseRouter *watermillHTTP.SSERouter, jwtAuthChecker *middleware.JWTAuthChecker) *Server {
	return &Server{
		Config:         config,
		Engine:         engine,
		Router:         router,
		sseRouter:      sseRouter,
		jwtAuthChecker: jwtAuthChecker,
	}
}

// RegisterRoutes method register all endpoints
func (s *Server) RegisterRoutes() {
	purchaseGroup := s.Engine.Group("/api/purchase")
	purchaseGroup.Use(s.jwtAuthChecker.JWTAuth())
	{
		purchaseGroup.POST("/", s.Router.PurchasingHandler.CreatePurchase)
		purchaseGroup.GET("/result", gin.WrapF(s.sseRouter.AddHandler(conf.PurchaseResultTopic, s.Router.PurchaseResultStreamHandler)))
	}
	go func() {
		err := s.sseRouter.Run(context.Background())
		if err != nil {
			log.Fatal(err)
		}
	}()

	<-s.sseRouter.Running()
}

// Run is a method for starting server
func (s *Server) Run() error {
	s.RegisterRoutes()
	addr := ":" + s.Config.Port
	s.Svr = &http.Server{
		Addr:    addr,
		Handler: s.Engine,
	}
	log.Infoln("listening on ", addr)
	err := s.Svr.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
