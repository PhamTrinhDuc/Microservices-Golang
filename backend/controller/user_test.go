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

type mockUserUsecase struct {
	RegisterFunc     func(ctx context.Context, req *domain.RegisterRequest) (*domain.User, error)
	AuthenticateFunc func(ctx context.Context, email, password string) (*domain.User, string, error)
	GetByIDFunc      func(ctx context.Context, id int) (*domain.User, error)
	LockUserFunc     func(ctx context.Context, id int, isLock bool) error
}

func (m *mockUserUsecase) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.User, error) {
	return m.RegisterFunc(ctx, req)
}

func (m *mockUserUsecase) Authenticate(ctx context.Context, email, password string) (*domain.User, string, error) {
	return m.AuthenticateFunc(ctx, email, password)
}

func (m *mockUserUsecase) GetByID(ctx context.Context, id int) (*domain.User, error) {
	return m.GetByIDFunc(ctx, id)
}

func (m *mockUserUsecase) LockUser(ctx context.Context, id int, isLock bool) error {
	return m.LockUserFunc(ctx, id, isLock)
}

func TestUserController_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           interface{}
		setupMock      func(m *mockUserUsecase)
		expectedStatus int
	}{
		{
			name: "success created",
			body: domain.RegisterRequest{
				FullName: "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMock: func(m *mockUserUsecase) {
				m.RegisterFunc = func(ctx context.Context, req *domain.RegisterRequest) (*domain.User, error) {
					return &domain.User{ID: 1, FullName: req.FullName, Email: req.Email}, nil
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid json body",
			body:           "invalid-json",
			setupMock:      func(m *mockUserUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "validation error - empty email",
			body: domain.RegisterRequest{
				FullName: "John Doe",
				Email:    "",
				Password: "password123",
			},
			setupMock:      func(m *mockUserUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "email already taken",
			body: domain.RegisterRequest{
				FullName: "John Doe",
				Email:    "taken@example.com",
				Password: "password123",
			},
			setupMock: func(m *mockUserUsecase) {
				m.RegisterFunc = func(ctx context.Context, req *domain.RegisterRequest) (*domain.User, error) {
					return nil, domain.ErrEmailTaken
				}
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "internal error",
			body: domain.RegisterRequest{
				FullName: "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMock: func(m *mockUserUsecase) {
				m.RegisterFunc = func(ctx context.Context, req *domain.RegisterRequest) (*domain.User, error) {
					return nil, errors.New("db disconnect")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			mockUC := &mockUserUsecase{}
			tt.setupMock(mockUC)

			ctl := controller.NewUserController(mockUC)
			r := gin.New()
			r.POST("/register", ctl.Register)

			var jsonBytes []byte
			var err error
			if strBody, ok := tt.body.(string); ok {
				jsonBytes = []byte(strBody)
			} else {
				jsonBytes, err = json.Marshal(tt.body)
				is.NoError(err)
			}

			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(jsonBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			is.Equal(tt.expectedStatus, w.Code)
		})
	}
}

func TestUserController_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           interface{}
		setupMock      func(m *mockUserUsecase)
		expectedStatus int
	}{
		{
			name: "success login",
			body: domain.LoginRequest{
				Email:    "user@example.com",
				Password: "password123",
			},
			setupMock: func(m *mockUserUsecase) {
				m.AuthenticateFunc = func(ctx context.Context, email, password string) (*domain.User, string, error) {
					return &domain.User{ID: 1, Email: email}, "mock-token", nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid json body",
			body:           "invalid-json",
			setupMock:      func(m *mockUserUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "validation error - invalid email format",
			body: domain.LoginRequest{
				Email:    "not-an-email",
				Password: "password123",
			},
			setupMock:      func(m *mockUserUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid credentials",
			body: domain.LoginRequest{
				Email:    "wrong@example.com",
				Password: "wrongpassword",
			},
			setupMock: func(m *mockUserUsecase) {
				m.AuthenticateFunc = func(ctx context.Context, email, password string) (*domain.User, string, error) {
					return nil, "", domain.ErrInvalidPassword
				}
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "account locked",
			body: domain.LoginRequest{
				Email:    "locked@example.com",
				Password: "password123",
			},
			setupMock: func(m *mockUserUsecase) {
				m.AuthenticateFunc = func(ctx context.Context, email, password string) (*domain.User, string, error) {
					return nil, "", domain.ErrLocked
				}
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			mockUC := &mockUserUsecase{}
			tt.setupMock(mockUC)

			ctl := controller.NewUserController(mockUC)
			r := gin.New()
			r.POST("/login", ctl.Login)

			var jsonBytes []byte
			var err error
			if strBody, ok := tt.body.(string); ok {
				jsonBytes = []byte(strBody)
			} else {
				jsonBytes, err = json.Marshal(tt.body)
				is.NoError(err)
			}

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(jsonBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			is.Equal(tt.expectedStatus, w.Code)
		})
	}
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
