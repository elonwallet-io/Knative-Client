package server

import (
	"backend/kubernetes_client/server/config"
	"backend/kubernetes_client/server/kubernetes"

	"github.com/labstack/echo/v4"
)

type Url struct {
	Url string `json:"url"`
}

type Server struct {
	Echo    *echo.Echo
	Clients *kubernetes.Kubernetes
	Config  config.Config
}

type User struct {
	Name string `json:"name"`
}
