package shipping

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"backend/domain"
)

type GHNClient struct {
	token    string
	shopID   int
	apiURL   string
	client   *http.Client
}

func NewGHNClient() *GHNClient {
	token := os.Getenv("GHN_TOKEN")
	shopIDStr := os.Getenv("GHN_SHOP_ID")
	apiURL := os.Getenv("GHN_API_URL")

	if apiURL == "" {
		apiURL = "https://dev-online-gateway.ghn.vn/shiip/public-api/v2"
	}

	shopID, _ := strconv.Atoi(shopIDStr)
	if shopID == 0 {
		shopID = 12345 // Sandbox default shop ID placeholder
	}

	return &GHNClient{
		token:  token,
		shopID: shopID,
		apiURL: apiURL,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

type GHNFeeRequest struct {
	FromDistrictID int    `json:"from_district_id"`
	ServiceTypeID  int    `json:"service_type_id"`
	ToDistrictID   int    `json:"to_district_id"`
	ToWardCode     string `json:"to_ward_code"`
	Weight         int    `json:"weight"`
}

type GHNFeeResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *struct {
		Total float64 `json:"total"`
	} `json:"data"`
}

type GHNOrderItem struct {
	Name     string `json:"name"`
	Code     string `json:"code"`
	Quantity int    `json:"quantity"`
	Price    int    `json:"price"`
}

type GHNCreateOrderRequest struct {
	PaymentTypeID  int            `json:"payment_type_id"` // 1: Shop pays (since shop collected from buyer)
	Note           string         `json:"note"`
	RequiredNote   string         `json:"required_note"` // KHONGCHOXEMHANG, CHOXEMHANGKHONGTHU, CHOXEMHANG
	ToName         string         `json:"to_name"`
	ToPhone        string         `json:"to_phone"`
	ToAddress      string         `json:"to_address"`
	ToWardCode     string         `json:"to_ward_code"`
	ToDistrictID   int            `json:"to_district_id"`
	Weight         int            `json:"weight"`
	ServiceTypeID  int            `json:"service_type_id"` // 2: E-commerce delivery
	Items          []GHNOrderItem `json:"items"`
}

type GHNCreateOrderResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *struct {
		OrderCode string `json:"order_code"` // Tracking code
	} `json:"data"`
}

// CalculateShippingFee calls GHN Sandbox to get shipping fee
func (c *GHNClient) CalculateShippingFee(toDistrictID int, toWardCode string, weight float64) (float64, error) {
	if c.token == "" {
		return 0, fmt.Errorf("GHN token not configured")
	}

	reqPayload := GHNFeeRequest{
		FromDistrictID: 1454, // Default Shop District ID (e.g. District 12, HCMC)
		ServiceTypeID:  2,    // standard E-commerce
		ToDistrictID:   toDistrictID,
		ToWardCode:     toWardCode,
		Weight:         int(weight),
	}

	bodyBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return 0, err
	}

	url := fmt.Sprintf("%s/shipping-order/fee", c.apiURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Token", c.token)
	req.Header.Set("ShopId", strconv.Itoa(c.shopID))

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("GHN fee API status code: %d", resp.StatusCode)
	}

	var ghnResp GHNFeeResponse
	if err := json.NewDecoder(resp.Body).Decode(&ghnResp); err != nil {
		return 0, err
	}

	if ghnResp.Code != 200 || ghnResp.Data == nil {
		return 0, fmt.Errorf("GHN fee API error: %s (code %d)", ghnResp.Message, ghnResp.Code)
	}

	return ghnResp.Data.Total, nil
}

// CreateShippingOrder calls GHN Sandbox to create shipment and returns tracking code
func (c *GHNClient) CreateShippingOrder(order *domain.Order, items []domain.OrderDetailResponse) (string, error) {
	if c.token == "" {
		return "", fmt.Errorf("GHN token not configured")
	}

	// Try to parse ward & district code from order address or fall back to mock sandbox values
	toDistrictID, _ := strconv.Atoi(order.ReceiverAddress) // if address_id wasn't integer, parse default mock
	if toDistrictID == 0 {
		toDistrictID = 1455 // Sandbox district fallback
	}
	toWardCode := "21012" // Sandbox ward fallback

	ghnItems := make([]GHNOrderItem, len(items))
	for i, item := range items {
		ghnItems[i] = GHNOrderItem{
			Name:     item.VariantName,
			Code:     item.SKU,
			Quantity: item.Quantity,
			Price:    int(item.UnitPrice),
		}
	}

	reqPayload := GHNCreateOrderRequest{
		PaymentTypeID: 1, // shop pays
		Note:          "Vui long giao gio hanh chinh",
		RequiredNote:  "KHONGCHOXEMHANG",
		ToName:        order.ReceiverName,
		ToPhone:       order.ReceiverPhone,
		ToAddress:     order.ReceiverAddress,
		ToWardCode:    toWardCode,
		ToDistrictID:  toDistrictID,
		Weight:        500 * len(items), // default 500g per item
		ServiceTypeID: 2,
		Items:         ghnItems,
	}

	bodyBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/shipping-order/create", c.apiURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Token", c.token)
	req.Header.Set("ShopId", strconv.Itoa(c.shopID))

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GHN create order API status code: %d", resp.StatusCode)
	}

	var ghnResp GHNCreateOrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&ghnResp); err != nil {
		return "", err
	}

	if ghnResp.Code != 200 || ghnResp.Data == nil {
		return "", fmt.Errorf("GHN create order API error: %s (code %d)", ghnResp.Message, ghnResp.Code)
	}

	return ghnResp.Data.OrderCode, nil
}
