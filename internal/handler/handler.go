package handler

import (
	"net/http"
	"strconv"
	"time"

	"g.pervovsky.ru/go-access-admin/internal/access"
	"g.pervovsky.ru/go-access-admin/internal/config"
	"g.pervovsky.ru/go-access-admin/internal/storage"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	DB     *storage.DB
	Config *config.Config
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/", h.RenderIndex)
	r.POST("/access", h.CreateAccess)
	r.GET("/access", h.GetAccessList)
	r.DELETE("/access/:id", h.DeleteAccess)
}

type AccessDTO struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	HtpasswdName string `json:"htpasswd_name"`
	ExpiresAt    string `json:"expires_at"`
	IsAdmin      bool   `json:"is_admin"`
	AccessLink   string `json:"access_link"`
}

func (h *Handler) RenderIndex(c *gin.Context) {
	paths := h.Config.HtpasswdPaths

	name := c.Query("htpasswd_name")

	var selected config.HtpasswdPath
	found := false
	for _, p := range paths {
		if name == "" || p.Name == name {
			selected = p
			found = true
			break
		}
	}

	// по умолчанию берём первый
	defaultName := ""
	if found {
		defaultName = selected.Name
	} else if len(paths) > 0 {
		defaultName = paths[0].Name
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"HtpasswdPaths": paths,
		"DefaultName":   defaultName,
	})
}

type CreateAccessRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	HtpasswdName    string `json:"htpasswd_name"`
	DurationMinutes int    `json:"duration_minutes"`
}

type CreateAccessResponse struct {
	AccessLink string `json:"access_link"`
	ExpiresAt  string `json:"expires_at"`
}

func (h *Handler) CreateAccess(c *gin.Context) {
	var req CreateAccessRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	// Найдём путь и шаблон
	var selected config.HtpasswdPath
	found := false
	for _, p := range h.Config.HtpasswdPaths {
		if p.Name == req.HtpasswdName {
			selected = p
			found = true
			break
		}
	}
	if !found {
		c.JSON(http.StatusBadRequest, gin.H{"error": "htpasswd_name not found"})
		return
	}

	// Запишем в htpasswd
	err := access.AddUser(selected.Path, req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update htpasswd"})
		return
	}

	expiresAt := time.Now().Add(time.Duration(req.DurationMinutes) * time.Minute)

	// Запишем в БД
	access := storage.Access{
		Username:     req.Username,
		Password:     req.Password,
		HtpasswdPath: selected.Path,
		ExpiresAt:    expiresAt,
		IsAdmin:      false,
	}

	_, err = h.DB.CreateAccess(access)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db insert error"})
		return
	}

	// Сгенерируем ссылку
	link := selected.URLTemplate
	link = replacePlaceholders(link, req.Username, req.Password)

	c.JSON(http.StatusOK, CreateAccessResponse{
		AccessLink: link,
		ExpiresAt:  expiresAt.Format(time.RFC3339),
	})
}

func replacePlaceholders(tmpl, user, pass string) string {
	tmpl = replaceAll(tmpl, "{user}", user)
	tmpl = replaceAll(tmpl, "{password}", pass)
	return tmpl
}

func replaceAll(s, old, new string) string {
	return string([]byte(
		(func(in string) string {
			return replaceOnce(in, old, new)
		})(s),
	))
}

func replaceOnce(s, old, new string) string {
	return string([]byte(
		(func() string {
			return s
		})(),
	))
}

func (h *Handler) GetAccessList(c *gin.Context) {
	name := c.Query("htpasswd_name")

	// если параметр не указан — берём первый
	var selected config.HtpasswdPath
	found := false
	for _, p := range h.Config.HtpasswdPaths {
		if name == "" || p.Name == name {
			selected = p
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or missing htpasswd_name"})
		return
	}

	records, err := h.DB.GetAccessesByPath(selected.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch accesses"})
		return
	}

	var result []AccessDTO
	for _, r := range records {
		result = append(result, AccessDTO{
			ID:           r.ID,
			Username:     r.Username,
			Password:     r.Password,
			HtpasswdName: selected.Name,
			ExpiresAt:    r.ExpiresAt.Format(time.RFC3339),
			IsAdmin:      r.IsAdmin,
			AccessLink:   replacePlaceholders(selected.URLTemplate, r.Username, r.Password),
		})
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) DeleteAccess(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}

	// Получим все доступы и найдём по id
	all, err := h.DB.GetAllAccesses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch accesses"})
		return
	}

	var toDelete *storage.Access
	for _, a := range all {
		if a.ID == id {
			toDelete = &a
			break
		}
	}

	if toDelete == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "access not found"})
		return
	}

	if toDelete.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot delete admin user"})
		return
	}

	// Удалим из htpasswd
	if err := access.RemoveUser(toDelete.HtpasswdPath, toDelete.Username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove from htpasswd"})
		return
	}

	// Удалим из БД
	if err := h.DB.DeleteAccess(int64(toDelete.ID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete from db"})
		return
	}

	c.Status(http.StatusNoContent)
}
