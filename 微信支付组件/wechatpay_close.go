package wechatpay

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CloseOrderRequest 关单请求参数
type CloseOrderRequest struct {
	OutTradeNo string `json:"out_trade_no"` // 商户订单号
	MchID      string `json:"mchid"`        // 商户号
}

// CloseOrder 关闭支付订单
// 参数:
//   - client: WechatPayClient实例
//   - req: 关单请求参数
// 返回:
//   - error: 错误信息
func CloseOrder(client *WechatPayClient, req *CloseOrderRequest) error {
	// 构建API路径
	path := fmt.Sprintf("/v3/pay/transactions/out-trade-no/%s/close", req.OutTradeNo)
	
	// 创建请求体
	requestBody := struct {
		MchID string `json:"mchid"`
	}{
		MchID: req.MchID,
	}
	
	// 发送POST请求
	resp, err := client.sendRequest(
		context.Background(),
		http.MethodPost,
		path,
		requestBody,
	)
	if err != nil {
		return fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode == http.StatusNoContent {
		return nil // 关单成功
	}

	// 处理错误响应
	body, _ := io.ReadAll(resp.Body)
	var wechatErr struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if json.Unmarshal(body, &wechatErr) == nil {
		return fmt.Errorf("wechatpay error[%d]: %s - %s", 
			resp.StatusCode, wechatErr.Code, wechatErr.Message)
	}
	return fmt.Errorf("unexpected status[%d]: %s", resp.StatusCode, string(body))
}
