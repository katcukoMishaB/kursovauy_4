package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *ProjectHandler) CreateInvitations(c *gin.Context) {
	var req CreateInvitationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	if len(req.Emails) == 0 && len(req.GroupIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Укажите хотя бы один email или группу"})
		return
	}
	n, err := h.service.CreateInvitations(c.Param("id"), c.GetHeader("X-User-ID"), req)
	if err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sent": n, "message": "Приглашения отправлены"})
}

func (h *ProjectHandler) ListMyInvitations(c *gin.Context) {
	invs, err := h.service.ListIncomingInvitations(c.GetHeader("X-User-ID"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить приглашения"})
		return
	}
	c.JSON(http.StatusOK, invs)
}

func (h *ProjectHandler) CountMyInvitations(c *gin.Context) {
	n, err := h.service.CountIncomingInvitations(c.GetHeader("X-User-ID"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить счётчик"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": n})
}

func (h *ProjectHandler) ListProjectInvitations(c *gin.Context) {
	invs, err := h.service.ListProjectInvitations(c.Param("id"), c.GetHeader("X-User-ID"))
	if err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, invs)
}

func (h *ProjectHandler) AcceptInvitation(c *gin.Context) {
	projectID, err := h.service.AcceptInvitation(c.Param("id"), c.GetHeader("X-User-ID"))
	if err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"project_id": projectID, "message": "Приглашение принято"})
}

func (h *ProjectHandler) RejectInvitation(c *gin.Context) {
	if err := h.service.RejectInvitation(c.Param("id"), c.GetHeader("X-User-ID")); err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Приглашение отклонено"})
}

func (h *ProjectHandler) CancelInvitation(c *gin.Context) {
	if err := h.service.CancelInvitation(c.Param("id"), c.GetHeader("X-User-ID")); err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Приглашение отменено"})
}
