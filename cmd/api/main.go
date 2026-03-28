// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name access_token
package main

import (
	"ValorantAPI/internal/api"
	"ValorantAPI/internal/deps"
	"ValorantAPI/internal/docs"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	initDeps := deps.New()
	defer initDeps.Logging.Sync()

	r := gin.New()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     initDeps.Cfg.CORS.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "Cookie"},
		AllowCredentials: true,
	}))
	r.Use(ginzap.Ginzap(initDeps.Logging.Desugar(), time.RFC3339, true))

	r.GET("/health", func(c *gin.Context) { c.String(200, "ok") })

	docs.SwaggerInfo.BasePath = "/"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api.LoadDefaultRouter(r, initDeps)

	apiSrv := &http.Server{
		Addr:              fmt.Sprintf("%v:%v", initDeps.Cfg.Server.Addr, initDeps.Cfg.Server.Port),
		Handler:           r,
		ReadTimeout:       time.Duration(initDeps.Cfg.Server.ReadTimeout) * time.Second,
		ReadHeaderTimeout: time.Duration(initDeps.Cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:      time.Duration(initDeps.Cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:       time.Duration(initDeps.Cfg.Server.IdleTimeout) * time.Second,
	}

	errCh := make(chan error, 1)
	go func() { errCh <- apiSrv.ListenAndServe() }()

	initDeps.Logging.Infow("server started", "addr", apiSrv.Addr)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		initDeps.Logging.Infow("signal received - shutting down",
			"signal", sig,
		)
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}

	ctxShut, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := apiSrv.Shutdown(ctxShut); err != nil {
		initDeps.Logging.Errorw("shutdown error", "err", err)
	}
	initDeps.Logging.Info("shutdown complete")
}
