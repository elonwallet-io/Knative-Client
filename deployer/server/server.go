package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"k8s.io/client-go/rest"
	"knative.dev/serving/pkg/client/clientset/versioned"
	servingv1 "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1"
)

type Url struct {
	Url string `json:"url"`
}

type Server struct {
	echo    *echo.Echo
	clients *ServingClients
	image   string
}

type ServingClients struct {
	Routes    servingv1.RouteInterface
	Configs   servingv1.ConfigurationInterface
	Revisions servingv1.RevisionInterface
	Services  servingv1.ServiceInterface
}

func newServingClients(cfg *rest.Config, namespace string) (*ServingClients, error) {
	cfg.QPS = 100
	cfg.Burst = 200
	cs, err := versioned.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &ServingClients{
		Configs:   cs.ServingV1().Configurations(namespace),
		Revisions: cs.ServingV1().Revisions(namespace),
		Routes:    cs.ServingV1().Routes(namespace),
		Services:  cs.ServingV1().Services(namespace),
	}, nil
}

type User struct {
	Name string `json:"name"`
}

func New() (*Server, error) {
	e := echo.New()
	e.Server.ReadTimeout = 5 * time.Second
	e.Server.WriteTimeout = 30 * time.Second
	e.Server.IdleTimeout = 120 * time.Second
	kubeconf, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconf: %w", err)
	}

	clients, err := newServingClients(kubeconf, "default")
	if err != nil {
		panic(err)
	}

	return &Server{
		echo:    echo.New(),
		clients: clients,
		image:   os.Getenv("IMAGENAME"),
	}, nil
}

func (s *Server) Run() (err error) {
	s.echo.POST("/enclaves", s.deployment)
	s.echo.DELETE("/enclaves/:id", s.deletion)
	port := os.Getenv("SERVERPORT")
	err = s.echo.Start(":" + port)
	if err == http.ErrServerClosed {
		err = nil
	}
	return
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return s.echo.Shutdown(ctx)
}

func (s *Server) deployment(c echo.Context) error {
	u := new(User)
	if err := c.Bind(u); err != nil {
		return err
	}
	if u.Name != "" {
		exists, _ := s.CheckIfServiceExists(c.Request().Context(), u.Name)
		fmt.Printf("service exists: %v", exists)
		if !exists {
			link, err := s.DeployContainer(c.Request().Context(), u.Name)
			if err != nil {
				return c.JSON(http.StatusOK, Url{Url: link})
			}
		}
		return c.String(http.StatusOK, "")
	} else {
		return c.String(http.StatusInternalServerError, "username can't be empty string")
	}
}

func (s *Server) deletion(c echo.Context) error {
	username := c.Param("id")
	fmt.Printf("Delete called for user %v \n", username)
	exists, _ := s.CheckIfServiceExists(c.Request().Context(), username)
	fmt.Printf("service exists: %v", exists)
	if exists {
		errors := s.clients.DeleteServiceForUser(c.Request().Context(), username)
		if len(errors) > 0 {
			return c.String(http.StatusInternalServerError, errors[0].Error())
		}
	}
	return c.String(http.StatusOK, "")
}
