package llm

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// GetAvailableModels retrieves a list of available LLMs from Ollama.
func GetAvailableModels(client *http.Client) ([]Model, error) {

	resp, err := client.Get("http://localhost:11434/api/tags")
	if err != nil {
		return nil, fmt.Errorf("failed to get available models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var res ollamaAvailableModels

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %v", err)
	}

	return res.Models, nil
}
