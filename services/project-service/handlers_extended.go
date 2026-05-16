package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *ProjectHandler) AdminCreateCategory(c *gin.Context) {
	var req struct{ Name string `json:"name"` }
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	id, err := h.service.AdminCreateCategory(req.Name)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Не удалось создать категорию"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "message": "Категория создана"})
}

func (h *ProjectHandler) AdminUpdateCategory(c *gin.Context) {
	var req struct{ Name string `json:"name"` }
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	if err := h.service.AdminUpdateCategory(c.Param("id"), req.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить категорию"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Категория обновлена"})
}

func (h *ProjectHandler) AdminDeleteCategory(c *gin.Context) {
	if err := h.service.AdminDeleteCategory(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить категорию"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Категория удалена"})
}

func (h *ProjectHandler) AdminCreateTag(c *gin.Context) {
	var req struct{ Name string `json:"name"` }
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	id, err := h.service.AdminCreateTag(req.Name)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Не удалось создать тег"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "message": "Тег создан"})
}

func (h *ProjectHandler) AdminUpdateTag(c *gin.Context) {
	var req struct{ Name string `json:"name"` }
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	if err := h.service.AdminUpdateTag(c.Param("id"), req.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить тег"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Тег обновлён"})
}

func (h *ProjectHandler) AdminDeleteTag(c *gin.Context) {
	if err := h.service.AdminDeleteTag(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить тег"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Тег удалён"})
}

func (h *ProjectHandler) ListGoals(c *gin.Context) {
	goals, err := h.service.ListGoals(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить цели"})
		return
	}
	c.JSON(http.StatusOK, goals)
}

func (h *ProjectHandler) CreateGoal(c *gin.Context) {
	var req CreateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	id, err := h.service.CreateGoal(c.Param("id"), c.GetHeader("X-User-ID"), req)
	if err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "message": "Цель добавлена"})
}

func (h *ProjectHandler) ToggleGoal(c *gin.Context) {
	var req struct {
		IsAchieved bool `json:"is_achieved"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	if err := h.service.ToggleGoal(c.Param("goalId"), c.GetHeader("X-User-ID"), req.IsAchieved); err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Цель обновлена"})
}

func (h *ProjectHandler) DeleteGoal(c *gin.Context) {
	if err := h.service.DeleteGoal(c.Param("goalId"), c.GetHeader("X-User-ID")); err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Цель удалена"})
}

func (h *ProjectHandler) ListRequiredSkills(c *gin.Context) {
	skills, err := h.service.ListRequiredSkills(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить навыки"})
		return
	}
	c.JSON(http.StatusOK, skills)
}

func (h *ProjectHandler) AddRequiredSkill(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	if err := h.service.AddRequiredSkill(c.Param("id"), c.GetHeader("X-User-ID"), req.Name); err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Навык добавлен"})
}

func (h *ProjectHandler) DeleteRequiredSkill(c *gin.Context) {
	if err := h.service.DeleteRequiredSkill(c.Param("id"), c.GetHeader("X-User-ID"), c.Param("name")); err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Навык удалён"})
}

func (h *ProjectHandler) AdminDeleteProject(c *gin.Context) {
	if err := h.service.AdminDeleteProject(c.Param("id")); err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Проект удалён"})
}

func (h *ProjectHandler) AdminRestoreProject(c *gin.Context) {
	if err := h.service.AdminRestoreProject(c.Param("id")); err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Проект восстановлен"})
}

func (h *ProjectHandler) CompleteProject(c *gin.Context) {
	if err := h.service.CompleteProject(c.Param("id"), c.GetHeader("X-User-ID")); err != nil {
		if err == ErrTasksOpen {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Нельзя завершить проект, есть незавершённые задачи"})
			return
		}
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Проект завершён"})
}

func (h *ProjectHandler) ListTagCatalog(c *gin.Context) {
	tags, err := h.service.ListTagCatalog()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить каталог"})
		return
	}
	c.JSON(http.StatusOK, tags)
}

func (h *ProjectHandler) AddTaskAssignee(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	if err := h.service.AddTaskAssignee(c.Param("id"), c.GetHeader("X-User-ID"), req.UserID); err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Исполнитель добавлен"})
}

func (h *ProjectHandler) RemoveTaskAssignee(c *gin.Context) {
	if err := h.service.RemoveTaskAssignee(c.Param("id"), c.GetHeader("X-User-ID"), c.Param("userId")); err != nil {
		code, msg := mapErr(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Исполнитель снят"})
}

func (h *ProjectHandler) ListTaskAssignees(c *gin.Context) {
	out, err := h.service.ListTaskAssignees(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить исполнителей"})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *ProjectHandler) Recommend(c *gin.Context) {
	recs, err := h.service.Recommend(c.GetHeader("X-User-ID"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось рассчитать рекомендации"})
		return
	}
	c.JSON(http.StatusOK, recs)
}
