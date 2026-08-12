package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type credentialsRequest struct {
	Email    string `json:"email" binding:"required"`
	Username string `json:"username"`
	Password string `json:"password" binding:"required"`
}

func RegisterRoutes(router *gin.Engine, service *Service) {
	router.POST("/api/auth/register", func(c *gin.Context) {
		var request credentialsRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "email and password are required"})
			return
		}

		user, err := service.Register(request.Email, request.Username, request.Password)
		if err != nil {
			if errors.Is(err, ErrEmailTaken) {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		respondWithSession(c, service, user)
	})

	router.POST("/api/auth/login", func(c *gin.Context) {
		var request credentialsRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "email and password are required"})
			return
		}

		user, err := service.Login(request.Email, request.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		respondWithSession(c, service, user)
	})
}

func respondWithSession(c *gin.Context, service *Service, user User) {
	token, err := service.CreateAccessToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create a session"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user": user, "accessToken": token})
}
