package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// JSONSchema represents a JSON schema as a JSON object
type JSONSchema map[string]interface{}

// Scan implements sql.Scanner
func (j *JSONSchema) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// Value implements driver.Valuer
func (j JSONSchema) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// Tool represents a tool/function definition
type Tool struct {
	ID           string     `gorm:"primaryKey" json:"id"`
	Name         string     `gorm:"uniqueIndex;not null" json:"name"`
	Description  string     `json:"description"`
	Category     string     `json:"category"` // search, generation, analysis, etc.
	InputSchema  JSONSchema `gorm:"type:jsonb" json:"input_schema"`
	OutputSchema JSONSchema `gorm:"type:jsonb" json:"output_schema"`
	Version      string     `json:"version"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// ToolPermission represents access control for tools
type ToolPermission struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ToolID    string    `gorm:"index" json:"tool_id"`
	AgentID   string    `gorm:"index" json:"agent_id"`
	Role      string    `json:"role"` // admin, user, agent
	CanUse    bool      `gorm:"default:true" json:"can_use"`
	CreatedAt time.Time `json:"created_at"`
}
