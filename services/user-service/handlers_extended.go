package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func (h *UserHandler) AdminListUsers(c *gin.Context) {
	users, err := h.service.repo.ListUsersWithRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить пользователей"})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) AdminCreateUser(c *gin.Context) {
	var req AdminCreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Имя, фамилия, email и пароль обязательны"})
		return
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось зашифровать пароль"})
		return
	}
	if !req.IsParticipant && !req.IsOrganizer && !req.IsAdmin {
		req.IsParticipant = true
	}
	id, err := h.service.repo.AdminCreateUser(req, string(hashed))
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique") {
			c.JSON(http.StatusConflict, gin.H{"error": "Email уже занят"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать пользователя"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "message": "Пользователь создан"})
}

func (h *UserHandler) AdminUpdateUser(c *gin.Context) {
	var req AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	var hashed *string
	if req.Password != "" {
		h, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось зашифровать пароль"})
			return
		}
		s := string(h)
		hashed = &s
	}
	if !req.IsParticipant && !req.IsOrganizer && !req.IsAdmin {
		req.IsParticipant = true
	}
	if err := h.service.repo.AdminUpdateUser(c.Param("id"), req, hashed); err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique") {
			c.JSON(http.StatusConflict, gin.H{"error": "Email уже занят"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить пользователя"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Пользователь обновлён"})
}

func (h *UserHandler) AdminDeleteUser(c *gin.Context) {
	currentID := c.GetHeader("X-User-ID")
	id := c.Param("id")
	if id == currentID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Нельзя удалить самого себя"})
		return
	}
	if err := h.service.repo.AdminDeleteUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить пользователя"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Пользователь удалён"})
}

func (h *UserHandler) AdminDashboard(c *gin.Context) {
	d, err := h.service.repo.Dashboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить дашборд"})
		return
	}
	c.JSON(http.StatusOK, d)
}

func (h *UserHandler) ListSkills(c *gin.Context) {
	skills, err := h.service.repo.ListSkills(c.GetHeader("X-User-ID"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить навыки"})
		return
	}
	c.JSON(http.StatusOK, skills)
}

func (h *UserHandler) AddSkill(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	if err := h.service.repo.AddSkill(c.GetHeader("X-User-ID"), req.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось добавить навык"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Навык добавлен"})
}

func (h *UserHandler) DeleteSkill(c *gin.Context) {
	if err := h.service.repo.DeleteSkill(c.GetHeader("X-User-ID"), c.Param("name")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить навык"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Навык удалён"})
}

func (h *UserHandler) ListInterests(c *gin.Context) {
	cats, err := h.service.repo.ListInterests(c.GetHeader("X-User-ID"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить интересы"})
		return
	}
	c.JSON(http.StatusOK, cats)
}

func (h *UserHandler) AddInterest(c *gin.Context) {
	var req struct {
		CategoryID string `json:"category_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.CategoryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	if err := h.service.repo.AddInterest(c.GetHeader("X-User-ID"), req.CategoryID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось добавить интерес"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Интерес добавлен"})
}

func (h *UserHandler) DeleteInterest(c *gin.Context) {
	if err := h.service.repo.DeleteInterest(c.GetHeader("X-User-ID"), c.Param("categoryId")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить интерес"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Интерес удалён"})
}
