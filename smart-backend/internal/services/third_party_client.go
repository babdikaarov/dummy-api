package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"ololo-gate/internal/config"
)

// ThirdPartyClient handles all communication with the third-party backend API
type ThirdPartyClient struct {
	baseURL string
	client  *http.Client
}

// LocationResponse represents a location from the third-party API with gates
type LocationResponse struct {
	ID      int                `json:"id"`
	Title   string             `json:"title"`
	Address string             `json:"address"`
	Logo    string             `json:"logo"`
	Gates   []GateResponse     `json:"gates"` // Gates should always be included in response
}

// LocationLiteDTO represents a lightweight location response without gates
type LocationLiteDTO struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Address string `json:"address"`
	Logo    string `json:"logo"`
}

// GateResponse represents a gate from the third-party API
type GateResponse struct {
	ID               int    `json:"id"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	LocationID       int    `json:"location_id"`
	IsOpen           bool   `json:"is_open"`
	GateIsHorizontal bool   `json:"gate_is_horizontal"`
}

// LocationAssignmentDTO represents a location with associated gate IDs
type LocationAssignmentDTO struct {
	LocationID int   `json:"locationId"`
	GateIds    []int `json:"gateIds"`
}

// UserLocationGateAssignmentDTO represents the request to assign user to locations/gates
// New nested structure: each location has its own array of gate IDs
type UserLocationGateAssignmentDTO struct {
	Phone     string                   `json:"phone"`
	Locations []LocationAssignmentDTO  `json:"locations"`
}

// NewThirdPartyClient creates a new instance of ThirdPartyClient
func NewThirdPartyClient() *ThirdPartyClient {
	return &ThirdPartyClient{
		baseURL: config.AppConfig.ThirdPartyAPIURL,
		client:  &http.Client{},
	}
}

// GetAllLocations fetches all locations with gates from the third-party API
func (c *ThirdPartyClient) GetAllLocations() ([]LocationResponse, error) {
	url := fmt.Sprintf("%s/locations", c.baseURL)
	resp, err := c.client.Get(url)
	if err != nil {
		log.Printf("Error calling third-party API GET %s: %v", url, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Third-party API returned status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("third-party API returned status code %d", resp.StatusCode)
	}

	// Read the entire body first for debugging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return nil, err
	}


	var locations []LocationResponse
	if err := json.Unmarshal(bodyBytes, &locations); err != nil {
		return nil, err
	}

	return locations, nil
}


// GetLocationsByPhone fetches all locations or locations filtered by phone from the third-party API
func (c *ThirdPartyClient) GetAllLocationsWithGates(phone string) ([]LocationResponse, error) {
	apiURL := fmt.Sprintf("%s/locations", c.baseURL)
	if phone != "" {
		// URL-encode the phone parameter to handle special characters like + sign
		apiURL = fmt.Sprintf("%s?phone=%s", apiURL, url.QueryEscape(phone))
	}

	resp, err := c.client.Get(apiURL)
	if err != nil {
		log.Printf("Error calling third-party API GET %s: %v", apiURL, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Third-party API returned status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("third-party API returned status code %d", resp.StatusCode)
	}

	var locations []LocationResponse
	if err := json.NewDecoder(resp.Body).Decode(&locations); err != nil {
		log.Printf("Error decoding locations response: %v", err)
		return nil, err
	}

	return locations, nil
}

// GetLocationsByPhone fetches locations accessible to a specific phone number
func (c *ThirdPartyClient) GetLocationsByPhone(phone string) ([]LocationLiteDTO, error) {
	url := fmt.Sprintf("%s/locations/by-phone/%s", c.baseURL, phone)
	resp, err := c.client.Get(url)
	if err != nil {
		log.Printf("Error calling third-party API GET %s: %v", url, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Third-party API returned status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("third-party API returned status code %d", resp.StatusCode)
	}

	var locations []LocationLiteDTO
	if err := json.NewDecoder(resp.Body).Decode(&locations); err != nil {
		log.Printf("Error decoding locations response: %v", err)
		return nil, err
	}

	return locations, nil
}

// GetGatesByPhoneAndLocation fetches gates accessible to a phone for a specific location
func (c *ThirdPartyClient) GetGatesByPhoneAndLocation(phone string, locationID int) ([]GateResponse, error) {
	url := fmt.Sprintf("%s/locations/by-phone/%s/%d", c.baseURL, phone, locationID)
	resp, err := c.client.Get(url)
	if err != nil {
		log.Printf("Error calling third-party API GET %s: %v", url, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Third-party API returned status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("third-party API returned status code %d", resp.StatusCode)
	}

	var gates []GateResponse
	if err := json.NewDecoder(resp.Body).Decode(&gates); err != nil {
		log.Printf("Error decoding gates response: %v", err)
		return nil, err
	}

	return gates, nil
}

// OpenGate sends a request to open a gate
func (c *ThirdPartyClient) OpenGate(gateID int) (bool, error) {
	log.Printf("[GATE_OPEN] Attempting to open gate ID: %d", gateID)
	url := fmt.Sprintf("%s/locations/%d/open", c.baseURL, gateID)
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		log.Printf("[GATE_OPEN] Error creating request for gate %d: %v", gateID, err)
		return false, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("[GATE_OPEN] Error calling third-party API for gate %d: %v", gateID, err)
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[GATE_OPEN] Third-party API returned status %d for gate %d: %s", resp.StatusCode, gateID, string(body))
		return false, fmt.Errorf("third-party API returned status code %d", resp.StatusCode)
	}

	var result bool
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[GATE_OPEN] Error decoding response for gate %d: %v", gateID, err)
		return false, err
	}

	log.Printf("[GATE_OPEN] Successfully opened gate ID: %d (result: %v)", gateID, result)
	return result, nil
}

// CloseGate sends a request to close a gate
func (c *ThirdPartyClient) CloseGate(gateID int) (bool, error) {
	log.Printf("[GATE_CLOSE] Attempting to close gate ID: %d", gateID)
	url := fmt.Sprintf("%s/locations/%d/close", c.baseURL, gateID)
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		log.Printf("[GATE_CLOSE] Error creating request for gate %d: %v", gateID, err)
		return false, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("[GATE_CLOSE] Error calling third-party API for gate %d: %v", gateID, err)
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[GATE_CLOSE] Third-party API returned status %d for gate %d: %s", resp.StatusCode, gateID, string(body))
		return false, fmt.Errorf("third-party API returned status code %d", resp.StatusCode)
	}

	var result bool
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[GATE_CLOSE] Error decoding response for gate %d: %v", gateID, err)
		return false, err
	}

	log.Printf("[GATE_CLOSE] Successfully closed gate ID: %d (result: %v)", gateID, result)
	return result, nil
}

// AssignUserToLocationsAndGates assigns a user (phone) to specific locations and gates
func (c *ThirdPartyClient) AssignUserToLocationsAndGates(assignment UserLocationGateAssignmentDTO) error {
	url := fmt.Sprintf("%s/locations/phone", c.baseURL)
	body, err := json.Marshal(assignment)
	if err != nil {
		log.Printf("Error marshaling assignment request: %v", err)
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error creating request to third-party API: %v", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("Error calling third-party API PUT %s: %v", url, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Third-party API returned status %d: %s", resp.StatusCode, string(body))
		return fmt.Errorf("third-party API returned status code %d", resp.StatusCode)
	}

	return nil
}
