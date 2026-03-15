package api

import (
	"context"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/cristianverduzco/nanodeploy/api/v1alpha1"
)

type Server struct {
	client client.Client
	router *gin.Engine
}

func NewServer(client client.Client) *Server {
	gin.SetMode(gin.ReleaseMode)
	s := &Server{
		client: client,
		router: gin.Default(),
	}
	s.router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		AllowCredentials: true,
	}))
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	api := s.router.Group("/api/v1")
	{
		api.GET("/services", s.listServices)
		api.GET("/services/:namespace/:name", s.getService)
		api.POST("/services", s.createService)
		api.DELETE("/services/:namespace/:name", s.deleteService)
	}
	s.router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

func (s *Server) Start(addr string) error {
	s.router.SetTrustedProxies(nil)
	return s.router.Run(addr)
}

// listServices returns all ManagedServices across all namespaces
func (s *Server) listServices(c *gin.Context) {
	list := &v1alpha1.ManagedServiceList{}
	if err := s.client.List(context.Background(), list, &client.ListOptions{}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"items": list.Items,
		"total": len(list.Items),
	})
}

// getService returns a single ManagedService by namespace and name
func (s *Server) getService(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	ms := &v1alpha1.ManagedService{}
	if err := s.client.Get(context.Background(), client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, ms); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}
	c.JSON(http.StatusOK, ms)
}

// createService provisions a new ManagedService
func (s *Server) createService(c *gin.Context) {
	var req struct {
		Name         string               `json:"name" binding:"required"`
		Namespace    string               `json:"namespace" binding:"required"`
		Type         v1alpha1.ServiceType `json:"type" binding:"required"`
		Version      string               `json:"version" binding:"required"`
		Replicas     int32                `json:"replicas"`
		StorageGB    int32                `json:"storageGB"`
		DatabaseName string               `json:"databaseName"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Replicas == 0 {
		req.Replicas = 1
	}
	if req.StorageGB == 0 {
		req.StorageGB = 5
	}

	ms := &v1alpha1.ManagedService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		Spec: v1alpha1.ManagedServiceSpec{
			Type:         req.Type,
			Version:      req.Version,
			Replicas:     req.Replicas,
			StorageGB:    req.StorageGB,
			DatabaseName: req.DatabaseName,
		},
	}

	if err := s.client.Create(context.Background(), ms); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ms)
}

// deleteService removes a ManagedService and all its owned resources
func (s *Server) deleteService(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	ms := &v1alpha1.ManagedService{}
	if err := s.client.Get(context.Background(), client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, ms); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}

	if err := s.client.Delete(context.Background(), ms); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "service deleted"})
}