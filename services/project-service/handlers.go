package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	service *ProjectService
}

func NewProjectHandler(service *ProjectService) *ProjectHandler {
	return &ProjectHandler{service: service}
}

func mapErr(err error) (int, string) {
	switch err {
	case ErrNotFound:
		return http.StatusNotFound, "Не найдено"
	case ErrForbidden:
		return http.StatusForbidden, "Недостаточно прав"
	case ErrAlreadyMember:
		return http.StatusConflict, "Вы уже участник этого проекта"
	case ErrAlreadyRequest:
		return http.StatusConflict, "У вас уже есть активная заявка на этот проект"
	case ErrInvalidRole:
		return http.StatusBadRequest, "Недопустимая роль"
	default:
		return http.StatusInternalServerError, "Внутренняя ошибка сервера"
	}
}

func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	id, err := h.service.CreateProject(c.GetHeader("X-User-ID"), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать проект"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "message": "Проект создан"})
}

func (h *ProjectHandler) ListProjects(c *gin.Context) {
	projects, err := h.service.ListProjects(c.Query("status"), c.Query("category_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить проекты"})
		return
	}
	c.JSON(http.StatusOK, projects)
}

func (h *ProjectHandler) ListMyProjects(c *gin.Context) {
	projects, err := h.service.ListMyProjects(c.GetHeader("X-User-ID"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить проекты"})
		return
	}
	c.JSON(http.StatusOK, projects)
}

func (h *ProjectHandler) GetProject(c *gin.Context) {
	p, err := h.service.GetProject(c.Param("id"))
	if err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	if err := h.service.UpdateProject(c.Param("id"), c.GetHeader("X-User-ID"), req); err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Проект обновлён"})
}

func (h *ProjectHandler) ArchiveProject(c *gin.Context) {
	if err := h.service.ArchiveProject(c.Param("id"), c.GetHeader("X-User-ID")); err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Проект архивирован"})
}

func (h *ProjectHandler) ListCategories(c *gin.Context) {
	categories, err := h.service.ListCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить категории"})
		return
	}
	c.JSON(http.StatusOK, categories)
}

func (h *ProjectHandler) CreateParticipationRequest(c *gin.Context) {
	var req ParticipationRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	id, err := h.service.CreateParticipationRequest(c.Param("id"), c.GetHeader("X-User-ID"), req)
	if err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "message": "Заявка отправлена"})
}

func (h *ProjectHandler) ListProjectRequests(c *gin.Context) {
	requests, err := h.service.ListProjectRequests(c.Param("id"), c.GetHeader("X-User-ID"))
	if err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, requests)
}

func (h *ProjectHandler) GetMyParticipationRequest(c *gin.Context) {
	req, err := h.service.GetMyParticipationRequest(c.Param("id"), c.GetHeader("X-User-ID"))
	if err == ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "Заявка не найдена"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных"})
		return
	}
	c.JSON(http.StatusOK, req)
}

func (h *ProjectHandler) ApproveRequest(c *gin.Context) {
	if err := h.service.ApproveRequest(c.Param("id"), c.GetHeader("X-User-ID")); err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Заявка одобрена"})
}

func (h *ProjectHandler) RejectRequest(c *gin.Context) {
	if err := h.service.RejectRequest(c.Param("id"), c.GetHeader("X-User-ID")); err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Заявка отклонена"})
}

func (h *ProjectHandler) ListParticipants(c *gin.Context) {
	participants, err := h.service.ListParticipants(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить участников"})
		return
	}
	c.JSON(http.StatusOK, participants)
}

func (h *ProjectHandler) UpdateParticipantRole(c *gin.Context) {
	var req struct {
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	if err := h.service.UpdateParticipantRole(c.Param("id"), c.Param("userId"), c.GetHeader("X-User-ID"), req.Role); err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Роль участника обновлена"})
}

func (h *ProjectHandler) AddTag(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	if err := h.service.AddTag(c.Param("id"), c.GetHeader("X-User-ID"), req.Name); err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Тег добавлен"})
}

func (h *ProjectHandler) GetProjectTags(c *gin.Context) {
	tags, err := h.service.GetProjectTags(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить теги"})
		return
	}
	c.JSON(http.StatusOK, tags)
}

func (h *ProjectHandler) DeleteTag(c *gin.Context) {
	if err := h.service.DeleteTag(c.Param("id"), c.GetHeader("X-User-ID"), c.Param("tagName")); err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Тег удалён"})
}
