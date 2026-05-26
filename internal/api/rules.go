package api

import (
	"net/http"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"

	"github.com/sunshow/siphongear/internal/rules"
	"github.com/sunshow/siphongear/internal/store/models"
)

type ruleIn struct {
	Name         string            `json:"name"`
	Enabled      bool              `json:"enabled"`
	Priority     int               `json:"priority"`
	IndicatorKey string            `json:"indicator_key"`
	TargetType   string            `json:"target_type"`
	TargetTags   []string          `json:"target_tags"`
	Conditions   []rules.Condition `json:"conditions"`
	Actions      []rules.Action    `json:"actions"`
}

type ruleOut struct {
	models.ThresholdRule
	TargetTagsArr []string          `json:"target_tags_arr"`
	Conditions    []rules.Condition `json:"conditions"`
	Actions       []rules.Action    `json:"actions"`
}

func toRuleOut(r models.ThresholdRule) ruleOut {
	conds, _ := rules.ParseConditions(r.ConditionJSON)
	acts, _ := rules.ParseActions(r.ActionJSON)
	return ruleOut{
		ThresholdRule: r,
		TargetTagsArr: rules.ParseTargetTags(r.TargetTags),
		Conditions:    conds,
		Actions:       acts,
	}
}

func validateRuleIn(in *ruleIn) error {
	in.Name = strings.TrimSpace(in.Name)
	in.IndicatorKey = strings.TrimSpace(in.IndicatorKey)
	in.TargetType = strings.TrimSpace(in.TargetType)
	if in.Name == "" {
		return errBadRequest("name is required")
	}
	if in.IndicatorKey == "" {
		return errBadRequest("indicator_key is required")
	}
	switch in.TargetType {
	case rules.TargetAll, rules.TargetTags:
	default:
		return errBadRequest("target_type must be 'all' or 'tags'")
	}
	if in.TargetType == rules.TargetTags && len(in.TargetTags) == 0 {
		return errBadRequest("target_tags is required when target_type is 'tags'")
	}
	if err := rules.ValidateConditions(in.Conditions); err != nil {
		return errBadRequest(err.Error())
	}
	if err := rules.ValidateActions(in.Actions); err != nil {
		return errBadRequest(err.Error())
	}
	return nil
}

type httpError struct {
	code int
	msg  string
}

func (e *httpError) Error() string { return e.msg }
func errBadRequest(m string) error { return &httpError{code: 400, msg: m} }

func writeRuleErr(c *gin.Context, err error) {
	if e, ok := err.(*httpError); ok {
		c.JSON(e.code, gin.H{"error": e.msg})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func applyRuleIn(in *ruleIn, row *models.ThresholdRule) error {
	tagsCSV := ""
	if in.TargetType == rules.TargetTags {
		clean := rules.ParseTargetTags(strings.Join(in.TargetTags, ","))
		tagsCSV = strings.Join(clean, ",")
	}
	condJSON, err := sonic.MarshalString(in.Conditions)
	if err != nil {
		return err
	}
	actJSON, err := sonic.MarshalString(in.Actions)
	if err != nil {
		return err
	}
	row.Name = in.Name
	row.Enabled = in.Enabled
	row.Priority = in.Priority
	row.IndicatorKey = in.IndicatorKey
	row.TargetType = in.TargetType
	row.TargetTags = tagsCSV
	row.ConditionJSON = condJSON
	row.ActionJSON = actJSON
	return nil
}

func (s *Server) listRules(c *gin.Context) {
	var rows []models.ThresholdRule
	if err := s.DB.Order("priority asc, id asc").Find(&rows).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	out := make([]ruleOut, 0, len(rows))
	for _, r := range rows {
		out = append(out, toRuleOut(r))
	}
	c.JSON(200, out)
}

func (s *Server) getRule(c *gin.Context) {
	var row models.ThresholdRule
	if err := s.DB.First(&row, c.Param("id")).Error; err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	c.JSON(200, toRuleOut(row))
}

func (s *Server) createRule(c *gin.Context) {
	var in ruleIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := validateRuleIn(&in); err != nil {
		writeRuleErr(c, err)
		return
	}
	var row models.ThresholdRule
	if err := applyRuleIn(&in, &row); err != nil {
		writeRuleErr(c, err)
		return
	}
	if err := s.DB.Create(&row).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, toRuleOut(row))
}

func (s *Server) updateRule(c *gin.Context) {
	var row models.ThresholdRule
	if err := s.DB.First(&row, c.Param("id")).Error; err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	var in ruleIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := validateRuleIn(&in); err != nil {
		writeRuleErr(c, err)
		return
	}
	if err := applyRuleIn(&in, &row); err != nil {
		writeRuleErr(c, err)
		return
	}
	if err := s.DB.Save(&row).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, toRuleOut(row))
}

func (s *Server) deleteRule(c *gin.Context) {
	if err := s.DB.Delete(&models.ThresholdRule{}, c.Param("id")).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}
