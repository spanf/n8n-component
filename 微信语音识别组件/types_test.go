package wechat_speech_recognition_test

import (
	"encoding/json"
	"testing"
	
	. "path/to/your/package" // 替换为实际包路径
)

func TestConstants(t *testing.T) {
	testCases := []struct {
		name     string
		actual   string
		expected string
	}{
		{"FormatAMR", FormatAMR, "amr"},
		{"FormatSpeex", FormatSpeex, "speex"},
		{"FormatMP3", FormatMP3, "mp3"},
		{"FormatWAV", FormatWAV, "wav"},
		{"LanguageZhCN", LanguageZhCN, "zh_CN"},
		{"LanguageEnUS", LanguageEnUS, "en_US"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.actual != tc.expected {
				t.Errorf("常量 %s 错误: 期望 %s, 实际 %s", tc.name, tc.expected, tc.actual)
			}
		})
	}
}

func TestUploadResponseJSON(t *testing.T) {
	testCases := []struct {
		name     string
		json     string
		expected string
	}{
		{"正常数据", `{"media_id":"MED123"}`, "MED123"},
		{"空数据", `{"media_id":""}`, ""},
		{"特殊字符", `{"media_id":"!@#$%^"}`, "!@#$%^"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var resp UploadResponse
			if err := json.Unmarshal([]byte(tc.json), &resp); err != nil {
				t.Fatalf("JSON解析失败: %v", err)
			}
			if resp.MediaID != tc.expected {
				t.Errorf("MediaID 错误: 期望 %s, 实际 %s", tc.expected, resp.MediaID)
			}
		})
	}
}

func TestRecognitionResponseJSON(t *testing.T) {
	testCases := []struct {
		name     string
		json     string
		expected string
	}{
		{"正常数据", `{"text":"你好世界"}`, "你好世界"},
		{"空数据", `{"text":""}`, ""},
		{"长文本", `{"text":"这是一段很长的语音识别文本内容..."}`, "这是一段很长的语音识别文本内容..."},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var resp RecognitionResponse
			if err := json.Unmarshal([]byte(tc.json), &resp); err != nil {
				t.Fatalf("JSON解析失败: %v", err)
			}
			if resp.Text != tc.expected {
				t.Errorf("Text 错误: 期望 %s, 实际 %s", tc.expected, resp.Text)
			}
		})
	}
}

func TestExtraJSONFields(t *testing.T) {
	t.Run("UploadResponse多余字段", func(t *testing.T) {
		jsonStr := `{"media_id":"ID123", "extra_field":"value"}`
		var resp UploadResponse
		if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
			t.Fatalf("解析包含多余字段的JSON失败: %v", err)
		}
		if resp.MediaID != "ID123" {
			t.Errorf("解析多余字段后 MediaID 错误: 期望 ID123, 实际 %s", resp.MediaID)
		}
	})

	t.Run("RecognitionResponse多余字段", func(t *testing.T) {
		jsonStr := `{"text":"测试文本", "unknown":123}`
		var resp RecognitionResponse
		if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
			t.Fatalf("解析包含多余字段的JSON失败: %v", err)
		}
		if resp.Text != "测试文本" {
			t.Errorf("解析多余字段后 Text 错误: 期望 '测试文本', 实际 '%s'", resp.Text)
		}
	})
}
