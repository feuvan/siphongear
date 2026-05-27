package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"

	"github.com/sunshow/siphongear/internal/notify"
	"github.com/sunshow/siphongear/internal/store/models"
)

type notifyChannelIn struct {
	Name    string         `json:"name"`
	Type    string         `json:"type"`
	Enabled *bool          `json:"enabled"`
	Notes   string         `json:"notes"`
	Payload map[string]any `json:"payload"`
}

func (s *Server) handleListNotifyTypes(c *gin.Context) {
	c.JSON(200, notify.ListMeta())
}

func (s *Server) listNotifyChannels(c *gin.Context) {
	var rows []models.NotificationChannel
	q := s.DB.Order("id desc")
	if t := c.Query("type"); t != "" {
		q = q.Where("type = ?", t)
	}
	if err := q.Find(&rows).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, rows)
}

func (s *Server) getNotifyChannel(c *gin.Context) {
	var row models.NotificationChannel
	if err := s.DB.First(&row, c.Param("id")).Error; err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	c.JSON(200, row)
}

func (s *Server) createNotifyChannel(c *gin.Context) {
	var in notifyChannelIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Type = strings.TrimSpace(in.Type)
	if in.Name == "" {
		c.JSON(400, gin.H{"error": "name is required"})
		return
	}
	if !notify.HasFactory(in.Type) {
		c.JSON(400, gin.H{"error": "unsupported notify type: " + in.Type})
		return
	}
	if _, err := notify.Build(in.Type, in.Payload); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	plain, _ := sonic.Marshal(in.Payload)
	enc, err := s.Cipher.EncryptString(string(plain))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	row := models.NotificationChannel{
		Name:    in.Name,
		Type:    in.Type,
		Enabled: enabled,
		Notes:   in.Notes,
		Payload: enc,
	}
	if err := s.DB.Create(&row).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, row)
}

func (s *Server) updateNotifyChannel(c *gin.Context) {
	var row models.NotificationChannel
	if err := s.DB.First(&row, c.Param("id")).Error; err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	var in notifyChannelIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Type = strings.TrimSpace(in.Type)
	if in.Name == "" {
		c.JSON(400, gin.H{"error": "name is required"})
		return
	}
	if in.Type == "" {
		in.Type = row.Type
	}
	if !notify.HasFactory(in.Type) {
		c.JSON(400, gin.H{"error": "unsupported notify type: " + in.Type})
		return
	}
	row.Name = in.Name
	row.Type = in.Type
	if in.Enabled != nil {
		row.Enabled = *in.Enabled
	}
	row.Notes = in.Notes
	if in.Payload != nil {
		if _, err := notify.Build(in.Type, in.Payload); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		plain, _ := sonic.Marshal(in.Payload)
		enc, err := s.Cipher.EncryptString(string(plain))
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		row.Payload = enc
	}
	if err := s.DB.Save(&row).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, row)
}

func (s *Server) deleteNotifyChannel(c *gin.Context) {
	if err := s.DB.Delete(&models.NotificationChannel{}, c.Param("id")).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

type notifyTestIn struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (s *Server) testNotifyChannel(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}
	var in notifyTestIn
	_ = c.ShouldBindJSON(&in)
	if strings.TrimSpace(in.Title) == "" {
		in.Title = "SiphonGear test notification"
	}
	if strings.TrimSpace(in.Body) == "" {
		in.Body = "This is a test message from SiphonGear."
	}
	if s.NotifyDispatcher == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "notify dispatcher unavailable"})
		return
	}
	err := s.NotifyDispatcher.SendOnce(c.Request.Context(), uint(id), notify.Message{
		Title:    in.Title,
		Body:     in.Body,
		Severity: notify.SeverityAlert,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (s *Server) listNotifyLogs(c *gin.Context) {
	var rows []models.NotificationLog
	q := s.DB.Order("id desc")
	if v := c.Query("channel_id"); v != "" {
		q = q.Where("channel_id = ?", v)
	}
	if v := c.Query("rule_id"); v != "" {
		q = q.Where("rule_id = ?", v)
	}
	if v := c.Query("collector_id"); v != "" {
		q = q.Where("collector_id = ?", v)
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	if err := q.Limit(limit).Find(&rows).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, rows)
}
