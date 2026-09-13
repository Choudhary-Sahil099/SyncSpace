package workspace

import (
	"fmt"
	"net/http"
	"strings"
	"syncspace/internal/auth"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, service *auth.Service) {
	router.GET("/api/workspaces", func(c *gin.Context) {
		user, ok := authenticate(c, service)
		if !ok {
			return
		}
		if err := service.EnsureDefaultWorkspace(user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not initialize workspace"})
			return
		}
		workspaces, err := service.ListWorkspaces(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load workspaces"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"workspaces": workspaces})
	})
	router.POST("/api/workspaces", func(c *gin.Context) {
		user, ok := authenticate(c, service)
		if !ok {
			return
		}
		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace payload"})
			return
		}
		w, err := service.CreateWorkspace(user.ID, req.Name, req.Description)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, w)
	})
	router.GET("/api/workspaces/:id", func(c *gin.Context) {
		user, ok := authenticate(c, service)
		if !ok {
			return
		}
		id, err := parseID(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace id"})
			return
		}
		w, err := service.GetWorkspace(user.ID, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
			return
		}
		c.JSON(http.StatusOK, w)
	})
	router.POST("/api/workspaces/:id/documents", func(c *gin.Context) {
		user, ok := authenticate(c, service)
		if !ok {
			return
		}
		id, err := parseID(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace id"})
			return
		}
		var req struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document payload"})
			return
		}
		d, err := service.CreateDocument(user.ID, id, req.Title, req.Description)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, d)
	})
	router.PATCH("/api/workspaces/:id/documents/:docId", func(c *gin.Context) {
		user, ok := authenticate(c, service)
		if !ok {
			return
		}
		wid, err := parseID(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace id"})
			return
		}
		docID, err := parseID(c.Param("docId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document id"})
			return
		}
		var req struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document payload"})
			return
		}
		d, err := service.UpdateDocument(user.ID, wid, docID, req.Title, req.Description)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, d)
	})
	router.DELETE("/api/workspaces/:id/documents/:docId", func(c *gin.Context) {
		user, ok := authenticate(c, service)
		if !ok {
			return
		}
		wid, err := parseID(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace id"})
			return
		}
		docID, err := parseID(c.Param("docId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document id"})
			return
		}
		if err := service.DeleteDocument(user.ID, wid, docID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	})
	router.GET("/api/workspaces/:id/tasks", func(c *gin.Context) {
		user, ok := authenticate(c, service)
		if !ok {
			return
		}
		wid, err := parseID(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace id"})
			return
		}
		tasks, err := service.ListTasks(user.ID, wid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"tasks": tasks})
	})
	router.POST("/api/workspaces/:id/tasks", func(c *gin.Context) {
		user, ok := authenticate(c, service)
		if !ok {
			return
		}
		wid, err := parseID(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace id"})
			return
		}
		var req struct {
			Title string `json:"title"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task payload"})
			return
		}
		task, err := service.CreateTask(user.ID, wid, req.Title)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, task)
	})
	router.PATCH("/api/workspaces/:id/tasks/:taskId", func(c *gin.Context) {
		user, ok := authenticate(c, service)
		if !ok {
			return
		}
		wid, err := parseID(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace id"})
			return
		}
		taskID, err := parseID(c.Param("taskId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
			return
		}
		var req struct {
			Completed bool `json:"completed"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task payload"})
			return
		}
		task, err := service.UpdateTask(user.ID, wid, taskID, req.Completed)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, task)
	})
	router.DELETE("/api/workspaces/:id/tasks/:taskId", func(c *gin.Context) {
		user, ok := authenticate(c, service)
		if !ok {
			return
		}
		wid, err := parseID(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace id"})
			return
		}
		taskID, err := parseID(c.Param("taskId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
			return
		}
		if err := service.DeleteTask(user.ID, wid, taskID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	})
	router.POST("/api/workspaces/:id/invites", func(c *gin.Context) {
		user, ok := authenticate(c, service)
		if !ok {
			return
		}
		id, err := parseID(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace id"})
			return
		}
		code, err := service.CreateInvite(user.ID, id)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "you cannot invite to this workspace"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"code": code})
	})
	router.POST("/api/workspaces/join", func(c *gin.Context) {
		user, ok := authenticate(c, service)
		if !ok {
			return
		}
		var req struct {
			Code string `json:"code"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Code) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invite code is required"})
			return
		}
		w, err := service.JoinWorkspace(user.ID, req.Code)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, w)
	})
}

func authenticate(c *gin.Context, service *auth.Service) (auth.User, bool) {
	token := c.GetHeader("Authorization")
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimSpace(strings.TrimPrefix(token, "Bearer "))
	} else {
		token = c.Query("access_token")
	}
	user, err := service.AuthenticateAccessToken(token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "a valid access token is required"})
		return auth.User{}, false
	}
	return user, true
}

func parseID(raw string) (int64, error) { var id int64; _, err := fmt.Sscan(raw, &id); return id, err }
