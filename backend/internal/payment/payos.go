package payment

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type PayOSClient struct {
	clientID    string
	apiKey      string
	checksumKey string
	apiURL      string
}

func NewPayOSClient() *PayOSClient {
	clientID := os.Getenv("PAYOS_CLIENT_ID")
	apiKey := os.Getenv("PAYOS_API_KEY")
	checksumKey := os.Getenv("PAYOS_CHECKSUM_KEY")
	apiURL := os.Getenv("PAYOS_API_URL")
	if apiURL == "" {
		apiURL = "https://api-sandbox.payos.vn"
	}

	return &PayOSClient{
		clientID:    clientID,
		apiKey:      apiKey,
		checksumKey: checksumKey,
		apiURL:      apiURL,
	}
}

type PayOSPaymentRequest struct {
	OrderCode   int64  `json:"orderCode"`
	Amount      int    `json:"amount"`
	Description string `json:"description"`
	CancelURL   string `json:"cancelUrl"`
	ReturnURL   string `json:"returnUrl"`
	Signature   string `json:"signature"`
}

type PayOSPaymentResponse struct {
	Code string `json:"code"`
	Desc string `json:"desc"`
	Data *struct {
		Bin           string `json:"bin"`
		AccountNumber string `json:"accountNumber"`
		AccountName   string `json:"accountName"`
		Amount        int    `json:"amount"`
		Description   string `json:"description"`
		OrderCode     int64  `json:"orderCode"`
		Currency      string `json:"currency"`
		PaymentLinkId string `json:"paymentLinkId"`
		Status        string `json:"status"`
		CheckoutURL   string `json:"checkoutUrl"`
		QrCode        string `json:"qrCode"`
	} `json:"data"`
}

type PayOSWebhookPayload struct {
	Code      string                 `json:"code"`
	Desc      string                 `json:"desc"`
	Data      map[string]interface{} `json:"data"`
	Signature string                 `json:"signature"`
}

// GenerateSignature generates PayOS signature for payment request
func (c *PayOSClient) GenerateSignature(req PayOSPaymentRequest) string {
	// Format: amount=xxx&cancelUrl=xxx&description=xxx&orderCode=xxx&returnUrl=xxx
	raw := fmt.Sprintf("amount=%d&cancelUrl=%s&description=%s&orderCode=%d&returnUrl=%s",
		req.Amount, req.CancelURL, req.Description, req.OrderCode, req.ReturnURL)

	h := hmac.New(sha256.New, []byte(c.checksumKey))
	h.Write([]byte(raw))
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateWebhookSignature generates checksum signature from data map
func (c *PayOSClient) GenerateWebhookSignature(data map[string]interface{}) string {
	var keys []string
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		val := data[k]
		if val == nil {
			parts = append(parts, fmt.Sprintf("%s=", k))
			continue
		}
		
		var valStr string
		switch v := val.(type) {
		case string:
			valStr = v
		case float64:
			// JSON numbers are parsed as float64
			valStr = strconv.FormatFloat(v, 'f', -1, 64)
		case int:
			valStr = strconv.Itoa(v)
		case int64:
			valStr = strconv.FormatInt(v, 10)
		case bool:
			valStr = strconv.FormatBool(v)
		default:
			valStr = fmt.Sprintf("%v", v)
		}
		parts = append(parts, fmt.Sprintf("%s=%s", k, valStr))
	}

	raw := strings.Join(parts, "&")
	h := hmac.New(sha256.New, []byte(c.checksumKey))
	h.Write([]byte(raw))
	return hex.EncodeToString(h.Sum(nil))
}

// CreatePaymentLink registers payment link with PayOS Sandbox
func (c *PayOSClient) CreatePaymentLink(orderCode int64, amount float64, description, returnURL, cancelURL string) (checkoutURL string, paymentCode string, err error) {
	// PayOS only supports integer amounts
	intAmount := int(amount)

	req := PayOSPaymentRequest{
		OrderCode:   orderCode,
		Amount:      intAmount,
		Description: description,
		CancelURL:   cancelURL,
		ReturnURL:   returnURL,
	}
	req.Signature = c.GenerateSignature(req)

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal payos request: %w", err)
	}

	url := fmt.Sprintf("%s/v2/payment-requests", c.apiURL)
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", "", fmt.Errorf("failed to create http request for payos: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-client-id", c.clientID)
	httpReq.Header.Set("x-api-key", c.apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", "", fmt.Errorf("failed to execute payos call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errData map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errData)
		return "", "", fmt.Errorf("payos API returned status %d: %v", resp.StatusCode, errData)
	}

	var apiResp PayOSPaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", "", fmt.Errorf("failed to decode payos response: %w", err)
	}

	if apiResp.Code != "00" || apiResp.Data == nil {
		return "", "", fmt.Errorf("payos payment link creation failed with code %s, desc: %s", apiResp.Code, apiResp.Desc)
	}

	// checkoutURL: URL to pay, paymentCode: paymentLinkId (used as unique payment reference ID in webhooks)
	return apiResp.Data.CheckoutURL, apiResp.Data.PaymentLinkId, nil
}

// VerifyWebhookSignature verifies webhook payload signature
func (c *PayOSClient) VerifyWebhookSignature(payload PayOSWebhookPayload) bool {
	expectedSig := c.GenerateWebhookSignature(payload.Data)
	return hmac.Equal([]byte(expectedSig), []byte(payload.Signature))
}
