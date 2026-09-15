package sms

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const moviderSMSEndpoint = "https://api.movider.co/v1/sms"

// MoviderClient sends SMS messages via the Movider API
// (https://developer.movider.co/reference/post_sms).
type MoviderClient struct {
	APIKey     string
	APISecret  string
	SenderName string
	HTTPClient *http.Client
}

func NewMoviderClient(apiKey, apiSecret, senderName string) *MoviderClient {
	return &MoviderClient{
		APIKey:     apiKey,
		APISecret:  apiSecret,
		SenderName: senderName,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// moviderSuccessResponse only decodes bad_phone_number_list - the sole field
// this client acts on. Movider's docs don't pin down the JSON type of the
// other success fields (e.g. remaining_balance), and a strictly-typed field
// never read shouldn't be able to fail the decode.
type moviderSuccessResponse struct {
	BadPhoneNumberList []struct {
		Number string `json:"number"`
		Msg    string `json:"msg"`
	} `json:"bad_phone_number_list"`
}

type moviderErrorResponse struct {
	Error struct {
		Code        int    `json:"code"`
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"error"`
}

func (m *MoviderClient) Send(ctx context.Context, to, message string) error {
	form := url.Values{
		"api_key":    {m.APIKey},
		"api_secret": {m.APISecret},
		"to":         {to},
		"text":       {message},
	}
	if m.SenderName != "" {
		form.Set("from", m.SenderName)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, moviderSMSEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := m.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp moviderErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("movider: %s (code %d): %s", errResp.Error.Name, errResp.Error.Code, errResp.Error.Description)
	}

	var okResp moviderSuccessResponse
	if err := json.NewDecoder(resp.Body).Decode(&okResp); err != nil {
		return err
	}
	if len(okResp.BadPhoneNumberList) > 0 {
		return fmt.Errorf("movider: rejected number %s: %s", okResp.BadPhoneNumberList[0].Number, okResp.BadPhoneNumberList[0].Msg)
	}

	return nil
}
