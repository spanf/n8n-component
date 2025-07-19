package sms

import (
	"encoding/json"
	"errors"
	"fmt"
)

type SMSClient struct {
	SecretId  string
	SecretKey string
	Endpoint  string
}

func buildSendSmsParams(templateId, phoneNumber, signName string, templateParams []string) map[string]string {
	params := make(map[string]string)
	params["TemplateId"] = templateId
	params["PhoneNumberSet.0"] = phoneNumber
	params["SignName"] = signName

	if len(templateParams) > 0 {
		for i, param := range templateParams {
			params[fmt.Sprintf("TemplateParamSet.%d", i)] = param
		}
	}

	return params
}

func (c *SMSClient) SendSMS(templateId, phoneNumber, signName string, templateParams []string) error {
	params := buildSendSmsParams(templateId, phoneNumber, signName, templateParams)
	params["Action"] = "SendSms"
	params["Version"] = "2021-01-11"
	params["Region"] = "ap-guangzhou"

	response, err := c.sendRequest(params)
	if err != nil {
		return err
	}

	var result struct {
		Response struct {
			Error struct {
				Code    string `json:"Code"`
				Message string `json:"Message"`
			} `json:"Error"`
		} `json:"Response"`
	}

	if err := json.Unmarshal(response, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Response.Error.Code != "" {
		return errors.New(result.Response.Error.Message)
	}

	return nil
}

func (c *SMSClient) sendRequest(params map[string]string) ([]byte, error) {
	return []byte{}, nil
}
