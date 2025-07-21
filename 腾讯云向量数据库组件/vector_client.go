package tencent_vector_db

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type VectorClient struct {
	secretId  string
	secretKey string
	endpoint  string
}

func NewVectorClient(secretId, secretKey, endpoint string) *VectorClient {
	return &VectorClient{
		secretId:  secretId,
		secretKey: secretKey,
		endpoint:  endpoint,
	}
}

func (c *VectorClient) internalGetSignature(method, path string, params map[string]string) string {
	timestamp := time.Now().Unix()
	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")
	service := "vector"
	region := c.extractRegion()

	canonicalURI := "/"
	if path != "" {
		canonicalURI = path
	}

	canonicalQuery := ""
	if len(params) > 0 {
		keys := make([]string, 0, len(params))
		for k := range params {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		queryParts := make([]string, 0, len(params))
		for _, k := range keys {
			key := url.QueryEscape(k)
			value := url.QueryEscape(params[k])
			queryParts = append(queryParts, key+"="+value)
		}
		canonicalQuery = strings.Join(queryParts, "&")
	}

	host := c.extractHost()
	canonicalHeaders := fmt.Sprintf("host:%s\n", host)
	signedHeaders := "host"

	payloadHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		method, canonicalURI, canonicalQuery, canonicalHeaders, signedHeaders, payloadHash)

	credentialScope := fmt.Sprintf("%s/%s/tc3_request", date, service)
	stringToSign := fmt.Sprintf("TC3-HMAC-SHA256\n%d\n%s\n%s",
		timestamp, credentialScope, hex.EncodeToString(hashSHA256([]byte(canonicalRequest))))

	secretDate := hmacSHA256([]byte("TC3"+c.secretKey), date)
	secretService := hmacSHA256(secretDate, service)
	secretSigning := hmacSHA256(secretService, "tc3_request")
	signature := hex.EncodeToString(hmacSHA256(secretSigning, stringToSign))

	return signature
}

func (c *VectorClient) internalSendRequest(method, path string, body interface{}) ([]byte, error) {
	u, err := url.Parse(c.endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint: %w", err)
	}
	u.Path = path

	var reqBody []byte
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
	}

	req, err := http.NewRequest(method, u.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	signature := c.internalGetSignature(method, path, nil)
	timestamp := time.Now().Unix()
	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")
	region := c.extractRegion()
	service := "vector"

	authHeader := fmt.Sprintf("TC3-HMAC-SHA256 Credential=%s/%s/%s/tc3_request, SignedHeaders=content-type;host, Signature=%s",
		c.secretId, date, region, signature)
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Host", u.Host)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return respBody, fmt.Errorf("API error: %s", resp.Status)
	}

	return respBody, nil
}

func (c *VectorClient) extractRegion() string {
	u, err := url.Parse(c.endpoint)
	if err != nil {
		return ""
	}
	hostParts := strings.Split(u.Hostname(), ".")
	if len(hostParts) < 3 {
		return ""
	}
	return hostParts[len(hostParts)-3]
}

func (c *VectorClient) extractHost() string {
	u, err := url.Parse(c.endpoint)
	if err != nil {
		return ""
	}
	return u.Hostname()
}

func hashSHA256(data []byte) []byte {
	hash := sha256.New()
	hash.Write(data)
	return hash.Sum(nil)
}

func hmacSHA256(key []byte, data string) []byte {
	hash := hmac.New(sha256.New, key)
	hash.Write([]byte(data))
	return hash.Sum(nil)
}
