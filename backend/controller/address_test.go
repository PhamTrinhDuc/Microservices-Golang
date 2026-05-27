package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/controller"
	"backend/domain"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockAddressUsecase struct {
	CreateFunc     func(ctx context.Context, userID int, req *domain.CreateAddressRequest) (*domain.Address, error)
	ListFunc       func(ctx context.Context, userID int) ([]*domain.Address, error)
	SetDefaultFunc func(ctx context.Context, userID, addressID int) error
	DeleteFunc     func(ctx context.Context, userID, addressID int) error
}

func (m *mockAddressUsecase) Create(ctx context.Context, userID int, req *domain.CreateAddressRequest) (*domain.Address, error) {
	return m.CreateFunc(ctx, userID, req)
}

func (m *mockAddressUsecase) List(ctx context.Context, userID int) ([]*domain.Address, error) {
	return m.ListFunc(ctx, userID)
}

func (m *mockAddressUsecase) SetDefault(ctx context.Context, userID, addressID int) error {
	return m.SetDefaultFunc(ctx, userID, addressID)
}

func (m *mockAddressUsecase) Delete(ctx context.Context, userID, addressID int) error {
	return m.DeleteFunc(ctx, userID, addressID)
}

func TestAddressController_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		injectUser     bool
		userIDValue    interface{}
		body           interface{}
		setupMock      func(m *mockAddressUsecase)
		expectedStatus int
	}{
		{
			name:        "success created",
			injectUser:  true,
			userIDValue: 42,
			body: domain.CreateAddressRequest{
				FullName:      "Home Address",
				Phone:         "0987654321",
				District:      "District 1",
				Province:      "HCM",
				Ward:          "Ward 1",
				DetailAddress: "123 Main St",
			},
			setupMock: func(m *mockAddressUsecase) {
				m.CreateFunc = func(ctx context.Context, userID int, req *domain.CreateAddressRequest) (*domain.Address, error) {
					return &domain.Address{
						ID:            1,
						UserID:        userID,
						FullName:      req.FullName,
						Phone:         req.Phone,
						District:      req.District,
						Province:      req.Province,
						Ward:          req.Ward,
						DetailAddress: req.DetailAddress,
					}, nil
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "unauthorized missing context user_id",
			injectUser:     false,
			body:           domain.CreateAddressRequest{},
			setupMock:      func(m *mockAddressUsecase) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:        "validation error - empty fields",
			injectUser:  true,
			userIDValue: 42,
			body: domain.CreateAddressRequest{
				FullName: "", // required
			},
			setupMock:      func(m *mockAddressUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "usecase internal error",
			injectUser:  true,
			userIDValue: 42,
			body: domain.CreateAddressRequest{
				FullName:      "Home Address",
				Phone:         "0987654321",
				District:      "District 1",
				Province:      "HCM",
				Ward:          "Ward 1",
				DetailAddress: "123 Main St",
			},
			setupMock: func(m *mockAddressUsecase) {
				m.CreateFunc = func(ctx context.Context, userID int, req *domain.CreateAddressRequest) (*domain.Address, error) {
					return nil, errors.New("db error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			mockUC := &mockAddressUsecase{}
			tt.setupMock(mockUC)

			ctl := controller.NewAddressController(mockUC)
			r := gin.New()
			r.POST("/addresses", func(ctx *gin.Context) {
				if tt.injectUser {
					ctx.Set("user_id", tt.userIDValue)
				}
				ctx.Next()
			}, ctl.Create)

			jsonBytes, err := json.Marshal(tt.body)
			is.NoError(err)

			req := httptest.NewRequest(http.MethodPost, "/addresses", bytes.NewReader(jsonBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			is.Equal(tt.expectedStatus, w.Code)
		})
	}
}

func TestAddressController_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		injectUser     bool
		userIDValue    interface{}
		setupMock      func(m *mockAddressUsecase)
		expectedStatus int
	}{
		{
			name:        "success list",
			injectUser:  true,
			userIDValue: 42,
			setupMock: func(m *mockAddressUsecase) {
				m.ListFunc = func(ctx context.Context, userID int) ([]*domain.Address, error) {
					return []*domain.Address{
						{ID: 1, UserID: userID, FullName: "Home"},
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized",
			injectUser:     false,
			setupMock:      func(m *mockAddressUsecase) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:        "internal error",
			injectUser:  true,
			userIDValue: 42,
			setupMock: func(m *mockAddressUsecase) {
				m.ListFunc = func(ctx context.Context, userID int) ([]*domain.Address, error) {
					return nil, errors.New("db error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			mockUC := &mockAddressUsecase{}
			tt.setupMock(mockUC)

			ctl := controller.NewAddressController(mockUC)
			r := gin.New()
			r.GET("/addresses", func(ctx *gin.Context) {
				if tt.injectUser {
					ctx.Set("user_id", tt.userIDValue)
				}
				ctx.Next()
			}, ctl.List)

			req := httptest.NewRequest(http.MethodGet, "/addresses", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			is.Equal(tt.expectedStatus, w.Code)
		})
	}
}

func TestAddressController_SetDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		injectUser     bool
		userIDValue    interface{}
		addressIDParam string
		setupMock      func(m *mockAddressUsecase)
		expectedStatus int
	}{
		{
			name:           "success set default",
			injectUser:     true,
			userIDValue:    42,
			addressIDParam: "10",
			setupMock: func(m *mockAddressUsecase) {
				m.SetDefaultFunc = func(ctx context.Context, userID, addressID int) error {
					return nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized missing context user_id",
			injectUser:     false,
			addressIDParam: "10",
			setupMock:      func(m *mockAddressUsecase) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid address id param",
			injectUser:     true,
			userIDValue:    42,
			addressIDParam: "abc",
			setupMock:      func(m *mockAddressUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "address not found",
			injectUser:     true,
			userIDValue:    42,
			addressIDParam: "99",
			setupMock: func(m *mockAddressUsecase) {
				m.SetDefaultFunc = func(ctx context.Context, userID, addressID int) error {
					return domain.ErrAddressNotFound
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "forbidden access address of other user",
			injectUser:     true,
			userIDValue:    42,
			addressIDParam: "11",
			setupMock: func(m *mockAddressUsecase) {
				m.SetDefaultFunc = func(ctx context.Context, userID, addressID int) error {
					return domain.ErrUnauthorized
				}
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			mockUC := &mockAddressUsecase{}
			tt.setupMock(mockUC)

			ctl := controller.NewAddressController(mockUC)
			r := gin.New()
			r.PUT("/addresses/:id/default", func(ctx *gin.Context) {
				if tt.injectUser {
					ctx.Set("user_id", tt.userIDValue)
				}
				ctx.Next()
			}, ctl.SetDefault)

			req := httptest.NewRequest(http.MethodPut, "/addresses/"+tt.addressIDParam+"/default", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			is.Equal(tt.expectedStatus, w.Code)
		})
	}
}

func TestAddressController_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		injectUser     bool
		userIDValue    interface{}
		addressIDParam string
		setupMock      func(m *mockAddressUsecase)
		expectedStatus int
	}{
		{
			name:           "success delete",
			injectUser:     true,
			userIDValue:    42,
			addressIDParam: "10",
			setupMock: func(m *mockAddressUsecase) {
				m.DeleteFunc = func(ctx context.Context, userID, addressID int) error {
					return nil
				}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "unauthorized missing context user_id",
			injectUser:     false,
			addressIDParam: "10",
			setupMock:      func(m *mockAddressUsecase) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid address id param",
			injectUser:     true,
			userIDValue:    42,
			addressIDParam: "abc",
			setupMock:      func(m *mockAddressUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "address not found",
			injectUser:     true,
			userIDValue:    42,
			addressIDParam: "99",
			setupMock: func(m *mockAddressUsecase) {
				m.DeleteFunc = func(ctx context.Context, userID, addressID int) error {
					return domain.ErrAddressNotFound
				}
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			mockUC := &mockAddressUsecase{}
			tt.setupMock(mockUC)

			ctl := controller.NewAddressController(mockUC)
			r := gin.New()
			r.DELETE("/addresses/:id", func(ctx *gin.Context) {
				if tt.injectUser {
					ctx.Set("user_id", tt.userIDValue)
				}
				ctx.Next()
			}, ctl.Delete)

			req := httptest.NewRequest(http.MethodDelete, "/addresses/"+tt.addressIDParam, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			is.Equal(tt.expectedStatus, w.Code)
		})
	}
}
