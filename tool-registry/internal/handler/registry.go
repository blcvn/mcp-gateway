package handler

import (
	"github.com/blcvn/backend/services/ba-tool-registry/internal/models"
	"gorm.io/gorm"
)

type ToolRegistry struct {
	db *gorm.DB
}

func NewToolRegistry(db *gorm.DB) *ToolRegistry {
	return &ToolRegistry{db: db}
}

// RegisterTool creates a new tool in the registry
func (r *ToolRegistry) RegisterTool(tool *models.Tool) error {
	return r.db.Create(tool).Error
}

// GetTool retrieves a tool by ID
func (r *ToolRegistry) GetTool(id string) (*models.Tool, error) {
	var tool models.Tool
	err := r.db.Where("id = ? AND is_active = ?", id, true).First(&tool).Error
	return &tool, err
}

// GetToolByName retrieves a tool by name
func (r *ToolRegistry) GetToolByName(name string) (*models.Tool, error) {
	var tool models.Tool
	err := r.db.Where("name = ? AND is_active = ?", name, true).First(&tool).Error
	return &tool, err
}

// ListTools returns all active tools
func (r *ToolRegistry) ListTools() ([]*models.Tool, error) {
	var tools []*models.Tool
	err := r.db.Where("is_active = ?", true).Find(&tools).Error
	return tools, err
}

// ListToolsByCategory returns tools in a specific category
func (r *ToolRegistry) ListToolsByCategory(category string) ([]*models.Tool, error) {
	var tools []*models.Tool
	err := r.db.Where("category = ? AND is_active = ?", category, true).Find(&tools).Error
	return tools, err
}

// UpdateTool updates tool metadata
func (r *ToolRegistry) UpdateTool(tool *models.Tool) error {
	return r.db.Save(tool).Error
}

// DeleteTool soft deletes a tool
func (r *ToolRegistry) DeleteTool(id string) error {
	return r.db.Model(&models.Tool{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

// GrantPermission grants tool access to an agent
func (r *ToolRegistry) GrantPermission(perm *models.ToolPermission) error {
	return r.db.Create(perm).Error
}

// CheckPermission checks if an agent can use a tool
func (r *ToolRegistry) CheckPermission(toolID, agentID string) (bool, error) {
	var perm models.ToolPermission
	err := r.db.Where("tool_id = ? AND agent_id = ? AND can_use = ?",
		toolID, agentID, true).First(&perm).Error

	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	return err == nil, err
}
