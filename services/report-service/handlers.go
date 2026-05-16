package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	service *ReportService
}

func NewReportHandler(service *ReportService) *ReportHandler {
	return &ReportHandler{service: service}
}

func mapErr(err error) (int, string) {
	switch err {
	case ErrNotFound:
		return http.StatusNotFound, "Не найдено"
	case ErrForbidden:
		return http.StatusForbidden, "Недостаточно прав"
	default:
		return http.StatusInternalServerError, "Не удалось сформировать отчёт"
	}
}

func (h *ReportHandler) ListUserActivity(c *gin.Context) {
	reports, err := h.service.ListUserActivity()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сформировать отчёт"})
		return
	}
	c.JSON(http.StatusOK, reports)
}

func (h *ReportHandler) GetUserActivity(c *gin.Context) {
	rep, err := h.service.GetUserActivity(c.Param("id"))
	if err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, rep)
}

func (h *ReportHandler) ListProjectEfficiency(c *gin.Context) {
	reports, err := h.service.ListProjectEfficiency()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сформировать отчёт"})
		return
	}
	c.JSON(http.StatusOK, reports)
}

func (h *ReportHandler) GetProjectEfficiency(c *gin.Context) {
	rep, err := h.service.GetProjectEfficiency(c.Param("id"))
	if err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, rep)
}

func (h *ReportHandler) GetSummary(c *gin.Context) {
	summary, err := h.service.GetSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сформировать сводку"})
		return
	}
	c.JSON(http.StatusOK, summary)
}
