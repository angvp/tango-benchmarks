package main

import (
	"github.com/gin-gonic/gin"

	"github.com/angvp/tango-benchmarks/apps/gin/store"
)

type server struct {
	q      *store.Queries
	tokens *tokens
}

func newServer(q *store.Queries, t *tokens) *server { return &server{q: q, tokens: t} }

func (s *server) routes() *gin.Engine {
	r := gin.Default()
	r.POST("/api/login", s.login)
	r.GET("/api/pets", s.listPets)
	r.GET("/api/pets/:id", s.getPet)
	authed := r.Group("/", s.requireToken)
	authed.POST("/api/pets", s.createPet)
	authed.PUT("/api/pets/:id", s.updatePet)
	authed.POST("/api/orders", s.createOrder)
	authed.GET("/api/orders/:id", s.getOrder)
	return r
}

func writeError(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": msg})
}
