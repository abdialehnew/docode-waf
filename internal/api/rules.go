package api

import (
	"net/http"

	"github.com/aleh/docode-waf/internal/models"
	"github.com/aleh/docode-waf/internal/services"
	"github.com/gin-gonic/gin"
)

type RulesHandler struct {
	service *services.RuleService
}

func NewRulesHandler(service *services.RuleService) *RulesHandler {
	return &RulesHandler{service: service}
}

func (h *RulesHandler) GetRules(c *gin.Context) {
	rules, err := h.service.GetAllRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rules"})
		return
	}
	c.JSON(http.StatusOK, rules)
}

func (h *RulesHandler) CreateRule(c *gin.Context) {
	var rule models.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreateRule(&rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create rule: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

func (h *RulesHandler) UpdateRule(c *gin.Context) {
	id := c.Param("id")
	var rule models.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule.ID = id

	if err := h.service.UpdateRule(&rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update rule: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rule updated successfully"})
}

func (h *RulesHandler) DeleteRule(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteRule(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete rule: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Rule deleted successfully"})
}

func (h *RulesHandler) ReorderRules(c *gin.Context) {
	// Optional: Bulk update priorities
	// Not implementing yet to keep it simple
	c.Status(http.StatusNotImplemented)
}
