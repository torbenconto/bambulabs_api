package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/torbenconto/bambulabs_api"
	"io"
	"net/http"
)

const (
	baseUrlCN = "https://api.bambulab.cn"
	baseUrlUS = "https://api.bambulab.com"
)

const (
	mqttHostCN = "cn.mqtt.bambulab.cn"
	mqttHostUS = "us.mqtt.bambulab.com"
)

type baseResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

type Client struct {
	region   Region
	email    string
	password string
	token    string
}

func NewClient(config *Config) *Client {
	return &Client{
		region:   config.Region,
		email:    config.Email,
		password: config.Password,
	}
}

func NewClientWithToken(config *Config, token string) *Client {
	client := NewClient(config)
	client.token = token
	return client
}

func (c *Client) getBaseUrl() string {
	if c.region == China {
		return baseUrlCN
	}

	return baseUrlUS
}

func (c *Client) getMqttHost() string {
	if c.region == China {
		return mqttHostCN
	}

	return mqttHostUS
}

type loginRequest struct {
	Email    string `json:"account"`
	Password string `json:"password"`
}
type loginResponse struct {
	Token     string `json:"accessToken"`
	LoginType string `json:"loginType"`
}

func (c *Client) Login() (string, error) {
	if c.token != "" {
		return c.token, nil
	}

	url := c.getBaseUrl() + "/user-service/user/login"

	body, err := json.Marshal(loginRequest{
		Email:    c.email,
		Password: c.password,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login failed: %s", response.Status)
	}

	body, err = io.ReadAll(response.Body)

	var loginResp loginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return "", err
	}

	if loginResp.LoginType == "verifyCode" {
		return "", nil
	}

	c.token = loginResp.Token

	return c.token, nil
}

type submitVerificationCodeRequest struct {
	Email string `json:"account"`
	Code  string `json:"code"`
}

func (c *Client) SubmitVerificationCode(code string) (string, error) {
	if c.token != "" {
		return c.token, nil
	}

	url := c.getBaseUrl() + "/user-service/user/login"

	body, err := json.Marshal(submitVerificationCodeRequest{
		Email: c.email,
		Code:  code,
	})

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login failed: %s", response.Status)
	}

	body, err = io.ReadAll(response.Body)

	var loginResp loginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return "", err
	}

	c.token = loginResp.Token

	return c.token, nil
}

type userInfoResponse struct {
	UserID string `json:"uid"`
}

func (c *Client) GetUserID() (string, error) {
	if c.token == "" {
		return "", fmt.Errorf("no token")
	}

	url := c.getBaseUrl() + "/design-user-service/my/preference"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get user id failed: %s", response.Status)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	var userInfoResp userInfoResponse
	if err := json.Unmarshal(body, &userInfoResp); err != nil {
		return "", err
	}

	return userInfoResp.UserID, nil
}

type getPrintersResponse struct {
	baseResponse
	Devices []struct {
		DevID          string `json:"dev_id"`
		Name           string `json:"name"`
		Online         bool   `json:"online"`
		PrintStatus    string `json:"print_status"`
		DevModelName   string `json:"dev_model_name"`
		DevProductName string `json:"dev_product_name"`
		DevAccessCode  string `json:"dev_access_code"`
	} `json:"devices"`
}

// GetPrintersAsPool returns a printer pool with all printers that are bound to the user.
// Please note that the ftp clients do not function over the cloud.
func (c *Client) GetPrintersAsPool() (*bambulabs_api.PrinterPool, error) {
	if c.token == "" {
		return &bambulabs_api.PrinterPool{}, fmt.Errorf("no token")
	}

	url := c.getBaseUrl() + "/iot-service/api/user/bind"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &bambulabs_api.PrinterPool{}, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return &bambulabs_api.PrinterPool{}, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return &bambulabs_api.PrinterPool{}, fmt.Errorf("get printers failed: %s", response.Status)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return &bambulabs_api.PrinterPool{}, err
	}

	var printersResp getPrintersResponse
	if err := json.Unmarshal(body, &printersResp); err != nil {
		return &bambulabs_api.PrinterPool{}, err
	}

	uid, err := c.GetUserID()
	if err != nil {
		return &bambulabs_api.PrinterPool{}, err
	}

	var printerConfigs []*bambulabs_api.PrinterConfig
	for _, device := range printersResp.Devices {
		printerConfigs = append(printerConfigs, &bambulabs_api.PrinterConfig{
			Host:         c.getMqttHost(),
			AccessCode:   c.token,
			SerialNumber: device.DevID,
			MqttUser:     uid,
			Mode:         bambulabs_api.CloudMode,
		})
	}

	pool := bambulabs_api.NewPrinterPool()
	for _, config := range printerConfigs {
		pool.AddPrinter(config)
	}

	return pool, nil
}
