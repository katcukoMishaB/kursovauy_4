package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	service *TaskService
}

func NewTaskHandler(service *TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

func mapErr(err error) (int, string) {
	switch err {
	case ErrNotFound:
		return http.StatusNotFound, "Не найдено"
	case ErrForbidden:
		return http.StatusForbidden, "Недостаточно прав"
	case ErrBadInput:
		return http.StatusBadRequest, "Некорректные данные"
	default:
		return http.StatusInternalServerError, "Внутренняя ошибка сервера"
	}
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	id, err := h.service.CreateTask(c.Param("project_id"), c.GetHeader("X-User-ID"), req)
	if err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "message": "Задача сохранена"})
}

func (h *TaskHandler) ListTasks(c *gin.Context) {
	tasks, err := h.service.ListTasks(c.Param("project_id"), c.Query("status"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить задачи"})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	t, err := h.service.GetTask(c.Param("id"))
	if err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	if err := h.service.UpdateTask(c.Param("id"), c.GetHeader("X-User-ID"), req); err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Задача сохранена"})
}

func (h *TaskHandler) AssignTask(c *gin.Context) {
	var req AssignTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	if err := h.service.AssignTask(c.Param("id"), c.GetHeader("X-User-ID"), req); err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Исполнитель назначен"})
}

func (h *TaskHandler) UpdateTaskStatus(c *gin.Context) {
	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	if err := h.service.UpdateTaskStatus(c.Param("id"), c.GetHeader("X-User-ID"), req.Status); err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Статус обновлён"})
}

func (h *TaskHandler) RateTask(c *gin.Context) {
	var req RateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	if err := h.service.RateTask(c.Param("id"), c.GetHeader("X-User-ID"), req.QualityRating); err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Оценка сохранена"})
}

func (h *TaskHandler) SetAttachment(c *gin.Context) {
	var req struct {
		AttachmentURL string `json:"attachment_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	if err := h.service.SetAttachment(c.Param("id"), c.GetHeader("X-User-ID"), req.AttachmentURL); err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Вложение обновлено"})
}

func (h *TaskHandler) AddComment(c *gin.Context) {
	var req CommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	id, err := h.service.AddComment(c.Param("id"), c.GetHeader("X-User-ID"), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось добавить комментарий"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "message": "Комментарий добавлен"})
}

func (h *TaskHandler) ListComments(c *gin.Context) {
	comments, err := h.service.ListComments(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить комментарии"})
		return
	}
	c.JSON(http.StatusOK, comments)
}
