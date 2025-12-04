package repository

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ElizCarvalho/fc-pos-golang-lab-weather-api-com-otel/service-b/internal/domain"
	"github.com/ElizCarvalho/fc-pos-golang-lab-weather-api-com-otel/service-b/internal/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockViaCEPClient struct {
	mock.Mock
}

func (m *MockViaCEPClient) GetLocationByZipcode(ctx context.Context, zipcode string) (*domain.Location, error) {
	args := m.Called(ctx, zipcode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Location), args.Error(1)
}

func TestViaCEPClientGetLocationByZipcode(t *testing.T) {
	tests := []struct {
		name           string
		zipcode        string
		mockResponse   *dto.ViaCEPResponse
		mockStatusCode int
		expected       *domain.Location
		expectedErr    error
	}{
		{
			name:    "success - valid zipcode",
			zipcode: "26140040",
			mockResponse: &dto.ViaCEPResponse{
				Localidade: "Belford Roxo",
				UF:         "RJ",
				Erro:       "",
			},
			mockStatusCode: http.StatusOK,
			expected: &domain.Location{
				City:  "Belford Roxo",
				State: "RJ",
			},
			expectedErr: nil,
		},
		{
			name:           "CEP não encontrado",
			zipcode:        "99999999",
			mockResponse:   &dto.ViaCEPResponse{Erro: "true"},
			mockStatusCode: http.StatusOK,
			expected:       nil,
			expectedErr:    domain.ErrZipcodeNotFound,
		},
		{
			name:           "CEP inválido",
			zipcode:        "123",
			mockResponse:   nil,
			mockStatusCode: http.StatusOK,
			expected:       nil,
			expectedErr:    domain.ErrInvalidZipcode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Criar servidor mock
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.mockStatusCode)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			// Criar cliente com URL do servidor mock
			client := NewViaCEPClient(server.URL)

			// Executar teste
			result, err := client.GetLocationByZipcode(context.Background(), tt.zipcode)

			// Verificar resultado
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
