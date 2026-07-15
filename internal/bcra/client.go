package bcra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	ErrInvalidCUIT = errors.New("invalid CUIT")
	ErrNotFound    = errors.New("BCRA debtor data not found")
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL: "https://api.bcra.gob.ar/CentralDeDeudores/v1.0",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) GetDebts(
	ctx context.Context,
	cuit string,
) (*DebtResponse, error) {
	cuit = strings.TrimSpace(cuit)

	if len(cuit) != 11 {
		return nil, ErrInvalidCUIT
	}

	endpoint := fmt.Sprintf(
		"%s/Deudas/%s",
		c.baseURL,
		cuit,
	)

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create BCRA request: %w", err)
	}

	request.Header.Set("Accept", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("execute BCRA request: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(
		io.LimitReader(response.Body, 5<<20),
	)
	if err != nil {
		return nil, fmt.Errorf("read BCRA response: %w", err)
	}

	if response.StatusCode == http.StatusNotFound {
		var errorResponse ErrorResponse

		if err := json.Unmarshal(body, &errorResponse); err != nil {
			return nil, ErrNotFound
		}

		message := "BCRA debtor data not found"

		if len(errorResponse.ErrorMessages) > 0 {
			message = errorResponse.ErrorMessages[0]
		}

		return nil, fmt.Errorf("%w: %s", ErrNotFound, message)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"BCRA returned HTTP status %d",
			response.StatusCode,
		)
	}

	var debtResponse DebtResponse

	if err := json.Unmarshal(body, &debtResponse); err != nil {
		return nil, fmt.Errorf("decode BCRA response: %w", err)
	}

	return &debtResponse, nil
}
