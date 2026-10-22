package llm

import "time"

type ollamaAvailableModels struct {
	Models []Model `json:"models"`
}

type Model struct {
	Name       string    `json:"name"`
	Model      string    `json:"model"`
	ModifiedAt time.Time `json:"modified_at"`
	Size       int64     `json:"size"`
	Digest     string    `json:"digest"`
	Details    struct {
		ParentModel       string `json:"parent_model"`
		Format            string `json:"format"`
		Family            string `json:"family"`
		ParameterSize     string `json:"parameter_size"`
		QuantizationLevel string `json:"quantization_level"`
		Runner            string `json:"runner"`
	} `json:"details"`
	Capabilities []string `json:"capabilities"`
}
