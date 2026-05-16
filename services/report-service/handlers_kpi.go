package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *ReportHandler) UserKPIs(c *gin.Context) {
	out, err := h.service.UserKPIsFiltered(
		c.Query("project_id"), c.Query("from"), c.Query("to"),
		c.Query("group_id"), c.Query("user_type"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось рассчитать KPI"})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *ReportHandler) GroupKPIs(c *gin.Context) {
	out, err := h.service.GroupKPIs(c.Query("from"), c.Query("to"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось рассчитать KPI по группам"})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *ReportHandler) UserTypeKPIs(c *gin.Context) {
	out, err := h.service.UserTypeKPIs(c.Query("from"), c.Query("to"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось рассчитать KPI по типам пользователей"})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *ReportHandler) ProjectKPIs(c *gin.Context) {
	out, err := h.service.ProjectKPIs(c.Query("from"), c.Query("to"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось рассчитать KPI"})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *ReportHandler) AdminDashboard(c *gin.Context) {
	out, err := h.service.DashboardData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить дашборд"})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *ReportHandler) ProjectDashboard(c *gin.Context) {
	role := c.GetHeader("X-User-Role")
	out, err := h.service.ProjectDashboardFiltered(
		c.Param("id"), c.GetHeader("X-User-ID"), role == "admin",
		c.Query("from"), c.Query("to"))
	if err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *ReportHandler) ProjectKPIExcel(c *gin.Context) {
	role := c.GetHeader("X-User-Role")
	data, filename, err := h.service.BuildProjectKPIExcel(
		c.Param("id"), c.GetHeader("X-User-ID"), role == "admin",
		c.Query("from"), c.Query("to"))
	if err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

func (h *ReportHandler) KPIExcel(c *gin.Context) {
	data, filename, err := h.service.BuildKPIExcelFiltered(
		c.Query("from"), c.Query("to"),
		c.Query("group_id"), c.Query("user_type"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сформировать отчёт"})
		return
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
