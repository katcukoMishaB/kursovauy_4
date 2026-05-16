package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service *UserService
}

func NewUserHandler(service *UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	id, err := h.service.Register(req)
	if err == ErrEmailExists {
		c.JSON(http.StatusConflict, gin.H{"error": "Email уже занят"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать пользователя"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "message": "Регистрация выполнена"})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	token, user, role, err := h.service.Login(req.Email, req.Password)
	switch err {
	case ErrInvalidLogin:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный email или пароль"})
		return
	case ErrUserDisabled:
		c.JSON(http.StatusForbidden, gin.H{"error": "Аккаунт пользователя заблокирован"})
		return
	case nil:
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось выполнить вход"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user, "role": role})
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	user, err := h.service.GetProfile(c.GetHeader("X-User-ID"))
	if err == ErrUserNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	if err := h.service.UpdateProfile(c.GetHeader("X-User-ID"), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить профиль"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Профиль обновлён"})
}

func (h *UserHandler) CreateOrganizerRequest(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}
	var req CreateOrganizerRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	id, err := h.service.CreateOrganizerRequest(userID, req)
	if err == ErrRequestExists {
		c.JSON(http.StatusConflict, gin.H{"error": "У вас уже есть активная заявка на рассмотрении"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать заявку"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "message": "Заявка подана"})
}

func (h *UserHandler) GetMyRequest(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	t := c.DefaultQuery("type", "organizer")
	if t != "organizer" && t != "teacher" {
		t = "organizer"
	}
	rq, err := h.service.GetMyLatestRequest(userID, t)
	if err == ErrRequestNotFound {
		c.JSON(http.StatusOK, nil)
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить заявку"})
		return
	}
	c.JSON(http.StatusOK, rq)
}

func (h *UserHandler) ListOrganizerRequests(c *gin.Context) {
	filter := c.Query("type")
	if filter != "" && filter != "organizer" && filter != "teacher" {
		filter = ""
	}
	requests, err := h.service.ListOrganizerRequests(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить заявки"})
		return
	}
	c.JSON(http.StatusOK, requests)
}

func (h *UserHandler) ApproveOrganizerRequest(c *gin.Context) {
	if err := h.service.ApproveOrganizerRequest(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось одобрить заявку"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Заявка одобрена"})
}

func (h *UserHandler) RejectOrganizerRequest(c *gin.Context) {
	if err := h.service.RejectOrganizerRequest(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось отклонить заявку"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Заявка отклонена"})
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	users, err := h.service.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить пользователей"})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	var req struct {
		Status bool `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	if err := h.service.UpdateUserStatus(c.Param("id"), req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить статус пользователя"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Статус пользователя обновлён"})
}

func (h *UserHandler) PublicListGroups(c *gin.Context) {
	groups, err := h.service.ListGroups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить группы"})
		return
	}
	c.JSON(http.StatusOK, groups)
}

func (h *UserHandler) AdminCreateGroup(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	id, err := h.service.CreateGroup(req.Name)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Группа уже существует или ошибка сохранения"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "message": "Группа создана"})
}

func (h *UserHandler) AdminUpdateGroup(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	if err := h.service.UpdateGroup(c.Param("id"), req.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить группу"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Группа обновлена"})
}

func (h *UserHandler) AdminDeleteGroup(c *gin.Context) {
	if err := h.service.DeleteGroup(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить группу"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Группа удалена"})
}
