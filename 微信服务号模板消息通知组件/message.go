package wechat_template_message

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type messageRequest struct {
	ToUser     string                 `json:"touser"`
	TemplateID string                 `json:"template_id"`
	URL        string                 `json:"url,omitempty"`
	MiniProgram *miniProgram          `json:"miniprogram,omitempty"`
	Data       map[string]interface{} `json:"data"`
}

type miniProgram struct {
	AppID    string `json:"appid"`
	PagePath string `json:"pagepath"`
}

type response struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	MsgID   int64  `json:"msgid"`
}

func SendTemplateMessage(accessToken, openID, templateID string, data map[string]interface{}, url string, miniprogram map[string]string) (int64, error) {
	body, err := buildMessageBody(openID, templateID, data, url, miniprogram)
	if err != nil {
		return 0, fmt.Errorf("build message body failed: %v", err)
	}

	respBody, err := sendMessageRequest(accessToken, body)
	if err != nil {
		return 0, fmt.Errorf("send request failed: %v", err)
	}

	var resp response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return 0, fmt.Errorf("parse response failed: %v", err)
	}

	if resp.ErrCode != 0 {
		return 0, fmt.Errorf("wechat api error: %d - %s", resp.ErrCode, resp.ErrMsg)
	}

	return resp.MsgID, nil
}

func buildMessageBody(openID, templateID string, data map[string]interface{}, url string, miniprogram map[string]string) ([]byte, error) {
	req := messageRequest{
		ToUser:     openID,
		TemplateID: templateID,
		Data:       data,
	}

	if url != "" {
		req.URL = url
	}

	if miniprogram != nil {
		if appID, ok := miniprogram["appid"]; ok {
			if pagePath, ok := miniprogram["pagepath"]; ok {
				req.MiniProgram = &miniProgram{
					AppID:    appID,
					PagePath: pagePath,
				}
			}
		}
	}

	return json.Marshal(req)
}

func sendMessageRequest(accessToken string, body []byte) ([]byte, error) {
	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=%s", accessToken)
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status: %d", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}
