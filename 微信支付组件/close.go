package wechatpay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Credential struct {
	ApiKey string
}

func CloseOrder(ctx context.Context, mchid string, outTradeNo string, credential *Credential) error {
	req, err := buildCloseOrderRequest(mchid, outTradeNo)
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)
	req.Header.Set("Authorization", "Bearer "+credential.ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return validateCloseOrderResponse(resp)
}

func buildCloseOrderRequest(mchid string, outTradeNo string) (*http.Request, error) {
	url := fmt.Sprintf("https://api.mch.weixin.qq.com/v3/pay/transactions/out-trade-no/%s/close", outTradeNo)
	
	payload := struct {
		MchID string `json:"mchid"`
	}{MchID: mchid}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func validateCloseOrderResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}
	return nil
}
