package contextguard

import (
	"charm.land/catwalk/pkg/catwalk"
	"charm.land/catwalk/pkg/embedded"
)

type ModelRegistryCrush struct {
	models map[string]catwalk.Model
}

// NewModelRegistryCrush - create a new model registry
func NewModelRegistryCrush() *ModelRegistryCrush {
	return &ModelRegistryCrush{}
}

// GetModels - get all models
func (m *ModelRegistryCrush) GetModels() map[string]catwalk.Model {
	if len(m.models) == 0 {
		for _, provider := range embedded.GetAll() {
			for _, model := range provider.Models {
				m.models[model.ID] = model
			}
		}
	}
	return m.models
}

// GetContextWindow - get context window from model
func (m *ModelRegistryCrush) GetContextWindow(modelID string) int {
	return int(m.models[modelID].ContextWindow)
}

// GetMaxOutputTokens - get max output tokens from model
func (m *ModelRegistryCrush) GetMaxOutputTokens(modelID string) int {
	return int(m.models[modelID].DefaultMaxTokens)
}
