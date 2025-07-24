package wechat_template_message_test

import (
	"testing"
	"path/to/your/package/wechat_template_message" // 替换为实际包路径
)

func TestConfigFields(t *testing.T) {
	tests := []struct {
		name      string
		appID     string
		appSecret string
		template  string
	}{
		{"AllFieldsSet", "wx123456", "sec7890", "TPL_001"},
		{"EmptyFields", "", "", ""},
		{"SpecialCharacters", "!@#$%^&*()", "+=[]{}|", "<>?~`"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := wechat_template_message.Config{
				AppID:      tt.appID,
				AppSecret:  tt.appSecret,
				TemplateID: tt.template,
			}

			if cfg.AppID != tt.appID {
				t.Errorf("AppID = %v, want %v", cfg.AppID, tt.appID)
			}
			if cfg.AppSecret != tt.appSecret {
				t.Errorf("AppSecret = %v, want %v", cfg.AppSecret, tt.appSecret)
			}
			if cfg.TemplateID != tt.template {
				t.Errorf("TemplateID = %v, want %v", cfg.TemplateID, tt.template)
			}
		})
	}
}

func TestConfigZeroValue(t *testing.T) {
	var cfg wechat_template_message.Config

	if cfg.AppID != "" {
		t.Errorf("Expected empty AppID, got %q", cfg.AppID)
	}
	if cfg.AppSecret != "" {
		t.Errorf("Expected empty AppSecret, got %q", cfg.AppSecret)
	}
	if cfg.TemplateID != "" {
		t.Errorf("Expected empty TemplateID, got %q", cfg.TemplateID)
	}
}
