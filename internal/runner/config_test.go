package runner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenCodeConfigUsesDeepseekV4Flash(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("..", "..", ".opencode", "opencode.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal(body, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Model != "opencode-go/deepseek-v4-flash" {
		t.Fatalf("model = %q, want opencode-go/deepseek-v4-flash", cfg.Model)
	}
}
