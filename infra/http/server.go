package http

import (
	"context"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	conf "github.com/minghsu0107/saga-purchase/config"
	"github.com/minghsu0107/saga-purchase/infra/http/middleware"
	"github.com/minghsu0107/saga-purchase/pkg"
	log "github.com/sirupsen/logrus"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	prommiddleware "github.com/slok/go-http-metrics/middleware"
	ginmiddleware "github.com/slok/go-http-metrics/middleware/gin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

// Server is the http wrapper
type Server struct {
	App            string
	Port           string
	Engine         *gin.Engine
	Router         *Router
	svr            *http.Server
	sseRouter      *pkg.SSERouter
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

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.LogMiddleware(config.Logger.ContextLogger))
	engine.Use(middleware.CORSMiddleware())

	mdlw := prommiddleware.New(prommiddleware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{
			Prefix: config.App,
		}),
	})
	engine.Use(ginmiddleware.Handler("", mdlw))
	return engine
}

// NewServer is the factory for server instance
func NewServer(config *conf.Config, engine *gin.Engine, router *Router, sseRouter *pkg.SSERouter, jwtAuthChecker *middleware.JWTAuthChecker) *Server {
	return &Server{
		App:            config.App,
		Port:           config.HTTPPort,
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
		purchaseGroup.POST("", s.Router.PurchasingHandler.CreatePurchase)
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
	addr := ":" + s.Port
	s.svr = &http.Server{
		Addr:    addr,
		Handler: newOtelHandler(s.Engine, s.App+"_http"),
	}
	log.Infoln("listening on ", addr)
	err := s.svr.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// GracefulStop the server
func (s *Server) GracefulStop(ctx context.Context) error {
	return s.svr.Shutdown(ctx)
}

func newOtelHandler(h http.Handler, operation string) http.Handler {
	httpOptions := []otelhttp.Option{
		otelhttp.WithTracerProvider(otel.GetTracerProvider()),
		otelhttp.WithPropagators(otel.GetTextMapPropagator()),
	}
	return otelhttp.NewHandler(h, operation, httpOptions...)
}
