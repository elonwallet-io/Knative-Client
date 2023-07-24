package server

import (
	"backend/kubernetes_client/server/config"
	"backend/kubernetes_client/server/kubernetes"
	"context"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
)

func New() (*Server, error) {
	e := echo.New()
	e.Server.ReadTimeout = 5 * time.Second
	e.Server.WriteTimeout = 120 * time.Second
	e.Server.IdleTimeout = 120 * time.Second
	return &Server{
		Echo:    echo.New(),
		Clients: kubernetes.CreateKubernetesClients(),
		Config:  config.CreateConfig(),
	}, nil
}

func (s *Server) Run() (err error) {
	s.Echo.POST("/enclaves", s.deployment)
	s.Echo.DELETE("/enclaves/:id", s.deletion)
	err = s.Echo.Start(":" + os.Getenv("SERVERPORT"))
	if err == http.ErrServerClosed {
		err = nil
	}
	return
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return s.Echo.Shutdown(ctx)
}
