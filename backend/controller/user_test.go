package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/controller"
	"backend/domain"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockUserUsecase struct {
	GetByIDFunc                   func(ctx context.Context, id int) (*domain.User, error)
	LockUserFunc                  func(ctx context.Context, id int, isLock bool) error
	ListFunc                      func(ctx context.Context, page, limit int, query string) ([]*domain.User, int, error)
}

func (m *mockUserUsecase) GetByID(ctx context.Context, id int) (*domain.User, error) {
	return m.GetByIDFunc(ctx, id)
}

func (m *mockUserUsecase) LockUser(ctx context.Context, id int, isLock bool) error {
	return m.LockUserFunc(ctx, id, isLock)
}

func (m *mockUserUsecase) List(ctx context.Context, page, limit int, query string) ([]*domain.User, int, error) {
	return m.ListFunc(ctx, page, limit, query)
}

func (m *mockUserUsecase) UpdateProfile(ctx context.Context, id int, req *domain.UpdateProfileRequest) (*domain.User, error) {
	return nil, nil
}

func (m *mockUserUsecase) UpdatePassword(ctx context.Context, id int, req *domain.UpdatePasswordRequest) error {
	return nil
}

func TestUserController_GetMe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		injectUser     bool
		userIDValue    interface{}
		setupMock      func(m *mockUserUsecase)
		expectedStatus int
	}{
		{
			name:        "success get profile",
			injectUser:  true,
			userIDValue: 42,
			setupMock: func(m *mockUserUsecase) {
				m.GetByIDFunc = func(ctx context.Context, id int) (*domain.User, error) {
					return &domain.User{ID: id, Email: "user@example.com"}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing context user_id",
			injectUser:     false,
			setupMock:      func(m *mockUserUsecase) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:        "user not found",
			injectUser:  true,
			userIDValue: 99,
			setupMock: func(m *mockUserUsecase) {
				m.GetByIDFunc = func(ctx context.Context, id int) (*domain.User, error) {
					return nil, domain.ErrUserNotFound
				}
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			mockUC := &mockUserUsecase{}
			tt.setupMock(mockUC)

			ctl := controller.NewUserController(mockUC)
			r := gin.New()

			// Middleware to inject context if required
			r.GET("/profile", func(ctx *gin.Context) {
				if tt.injectUser {
					ctx.Set("user_id", tt.userIDValue)
				}
				ctx.Next()
			}, ctl.GetMe)

			req := httptest.NewRequest(http.MethodGet, "/profile", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			is.Equal(tt.expectedStatus, w.Code)
		})
	}
}

func TestUserController_LockUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userIDParam    string
		body           interface{}
		setupMock      func(m *mockUserUsecase)
		expectedStatus int
	}{
		{
			name:        "success lock",
			userIDParam: "42",
			body:        gin.H{"is_lock": true},
			setupMock: func(m *mockUserUsecase) {
				m.LockUserFunc = func(ctx context.Context, id int, isLock bool) error {
					return nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid user id param",
			userIDParam:    "invalid",
			body:           gin.H{"is_lock": true},
			setupMock:      func(m *mockUserUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "user not found",
			userIDParam: "99",
			body:        gin.H{"is_lock": true},
			setupMock: func(m *mockUserUsecase) {
				m.LockUserFunc = func(ctx context.Context, id int, isLock bool) error {
					return domain.ErrUserNotFound
				}
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			mockUC := &mockUserUsecase{}
			tt.setupMock(mockUC)

			ctl := controller.NewUserController(mockUC)
			r := gin.New()
			r.PUT("/users/:id/lock", ctl.LockUser)

			jsonBytes, err := json.Marshal(tt.body)
			is.NoError(err)

			req := httptest.NewRequest(http.MethodPut, "/users/"+tt.userIDParam+"/lock", bytes.NewReader(jsonBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			is.Equal(tt.expectedStatus, w.Code)
		})
	}
}


