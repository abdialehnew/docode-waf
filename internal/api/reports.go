package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aleh/docode-waf/internal/services"
	"github.com/gin-gonic/gin"
)

type ReportsHandler struct {
	service *services.ReportingService
}

func NewReportsHandler(service *services.ReportingService) *ReportsHandler {
	return &ReportsHandler{service: service}
}

func (h *ReportsHandler) GenerateReport(c *gin.Context) {
	format := c.Query("format") // pdf or docx
	startStr := c.Query("start")
	endStr := c.Query("end")
	lang := c.Query("lang")
	if lang == "" {
		lang = "en"
	}

	if format != "pdf" && format != "docx" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid format. Supported: pdf, docx"})
		return
	}

	// Parse dates
	layout := "2006-01-02"
	start, err := time.Parse(layout, startStr)
	if err != nil {
		// Default to last 7 days
		start = time.Now().AddDate(0, 0, -7)
	}

	end, err := time.Parse(layout, endStr)
	if err != nil {
		// Default to now
		end = time.Now()
	}

	// Adjust end to end of day
	end = end.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	// Fetch data
	data, err := h.service.GenerateReportData(start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate report data: " + err.Error()})
		return
	}

	var content []byte
	var filename string
	var contentType string

	if format == "pdf" {
		content, err = h.service.GeneratePDF(data, lang)
		filename = fmt.Sprintf("waf-report-%s.pdf", time.Now().Format("20060102-150405"))
		contentType = "application/pdf"
	} else {
		content, err = h.service.GenerateDOCX(data, lang)
		filename = fmt.Sprintf("waf-report-%s.docx", time.Now().Format("20060102-150405"))
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate document: " + err.Error()})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, contentType, content)
}
