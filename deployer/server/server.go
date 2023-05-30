package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"knative.dev/serving/pkg/client/clientset/versioned"
	test "knative.dev/serving/test"
)

type Url struct {
	Url string `json:"url"`
}

type Server struct {
	Echo    *echo.Echo
	Clients *Kubernetes
	Config  Config
}

type Kubernetes struct {
	Kubernetes *kubernetes.Clientset
	Knative    *test.ServingClients
}

func newServingClients(cfg *rest.Config, namespace string) (*Kubernetes, error) {
	cfg.QPS = 100
	cfg.Burst = 200
	cs, err := versioned.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return &Kubernetes{
		Kubernetes: clientset,
		Knative: &test.ServingClients{
			Configs:   cs.ServingV1().Configurations(namespace),
			Revisions: cs.ServingV1().Revisions(namespace),
			Routes:    cs.ServingV1().Routes(namespace),
			Services:  cs.ServingV1().Services(namespace),
		},
	}, nil
}

type Config struct {
	Image         string
	FRONTEND_URL  string
	FRONTEND_HOST string
	BACKEND_URL   string
	SGX_ACTIVATE  bool
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
	sgx_activate, err := strconv.ParseBool(os.Getenv("SGX_ACTIVATE"))
	if err != nil {
		panic(err)
	}

	return &Server{
		Echo:    echo.New(),
		Clients: clients,
		Config: Config{
			Image:         os.Getenv("IMAGENAME"),
			FRONTEND_URL:  os.Getenv("FRONTEND_URL"),
			BACKEND_URL:   os.Getenv("BACKEND_URL"),
			FRONTEND_HOST: os.Getenv("FRONTEND_HOST"),
			SGX_ACTIVATE:  sgx_activate,
		},
	}, nil
}

func (s *Server) Run() (err error) {
	s.Echo.POST("/enclaves", s.deployment)
	s.Echo.DELETE("/enclaves/:id", s.deletion)
	port := os.Getenv("SERVERPORT")
	err = s.Echo.Start(":" + port)
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

func (s *Server) deployment(c echo.Context) error {
	u := new(User)
	if err := c.Bind(u); err != nil {
		return err
	}
	if u.Name != "" {
		fmt.Println("checking if service " + u.Name + "exists... ")
		route, _ := s.CheckIfServiceExists(c.Request().Context(), u.Name)
		fmt.Println("service exists: " + route)
		if route == "" {
			fmt.Println("deploying container..")
			var err error
			route, err = s.DeployContainer(c.Request().Context(), u.Name)
			if err != nil {
				fmt.Println("error: " + err.Error())
				return c.JSON(http.StatusInternalServerError, err.Error())
			}
		}
		fmt.Println(route)
		return c.JSON(http.StatusOK, Url{Url: route})

	} else {
		return c.String(http.StatusInternalServerError, "username can't be empty string")
	}
}

func (s *Server) deletion(c echo.Context) error {
	username := c.Param("id")
	fmt.Printf("Delete called for user %v \n", username)
	route, _ := s.CheckIfServiceExists(c.Request().Context(), username)
	fmt.Printf("service exists: %v", route)
	if route != "" {
		errors := s.Clients.DeleteServiceForUser(c.Request().Context(), username)
		if len(errors) > 0 {
			return c.String(http.StatusInternalServerError, errors[0].Error())
		}
	}
	return c.String(http.StatusOK, "")
}
