package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"agentflow/internal/repo"

	"github.com/gin-gonic/gin"
)

// platformLanguage reads (administrators only) or saves the platform-wide
// interface language under 平台管理.
func (h *usersHandlers) platformLanguage(c *gin.Context) {
	switch c.Request.Method {
	case http.MethodGet:
		if !h.isPlatformAdmin(c.Request) {
			c.JSON(http.StatusForbidden, gin.H{"error": "只有平台管理员可以查看界面语言设置"})
			return
		}
		c.JSON(http.StatusOK, h.repo.PlatformInterfaceLanguageStatus())
	case http.MethodPut:
		if !h.platformAdminWrite(c) {
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<10)
		var input struct {
			Language *string `json:"language"`
		}
		if c.ShouldBindJSON(&input) != nil || input.Language == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid interface language request"})
			return
		}
		var actorID int64
		if current, ok := h.requestUser(c.Request, true); ok {
			actorID = current.ID
		}
		err := h.repo.SetPlatformInterfaceLanguage(*input.Language, actorID)
		if errors.Is(err, repo.ErrInvalidInterfaceLanguage) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "界面语言只能是由用户切换（留空）、英文（en）或中文（zh）"})
			return
		}
		if err != nil {
			respond(c, err)
			return
		}
		c.JSON(http.StatusOK, h.repo.PlatformInterfaceLanguageStatus())
	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	}
}

// platformLanguageScript tells every page, signed in or not, which language
// the platform locks the interface to before it paints. It carries only the
// language value ("", "en" or "zh").
func (h *usersHandlers) platformLanguageScript(c *gin.Context) {
	value, _ := json.Marshal(h.repo.PlatformInterfaceLanguage())
	c.Header("Content-Type", "text/javascript; charset=utf-8")
	c.Header("Cache-Control", "no-store")
	c.String(http.StatusOK, "window.LF_LANG_POLICY="+string(value)+";\n")
}
