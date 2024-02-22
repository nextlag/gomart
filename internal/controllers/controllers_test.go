package controllers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nextlag/gomart/internal/config"
	"github.com/nextlag/gomart/internal/controllers/mocks"
	"github.com/nextlag/gomart/internal/usecase"
	"github.com/nextlag/gomart/pkg/logger/slogpretty"
)

func controller(t *testing.T) (*Controller, *mocks.MockUseCase, *usecase.UseCase) {
	t.Helper()
	log := slogpretty.SetupLogger(config.ProjectRoot)
	var cfg config.HTTPServer
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	repo := mocks.NewMockUseCase(mockCtl)
	db := usecase.NewMockRepository(mockCtl)
	uc := usecase.New(db, cfg)

	controller := New(repo, log)
	return controller, repo, uc
}

func TestRegistrationHandler(t *testing.T) {
	type want struct {
		statusCode int
	}

	tests := []struct {
		name string
		want want
		body string
	}{
		{
			name: "Registration success",
			want: want{statusCode: http.StatusOK},
			body: `{"login": "test", "password": "12345"}`,
		},
		{
			name: "Duplicate login",
			want: want{statusCode: http.StatusConflict},
			body: `{"login": "duplicate", "password": "12345"}`,
		},
		{
			name: "Internal server error",
			want: want{statusCode: http.StatusInternalServerError},
			body: `{"login": "error", "password": "error"}`,
		},
		{
			name: "Empty password",
			want: want{statusCode: http.StatusOK},
			body: `{"login": "one"}`,
		},
		{
			name: "Empty login",
			want: want{statusCode: http.StatusBadRequest},
			body: `{"password": "12345"}`,
		},
		{
			name: "Invalid request",
			want: want{statusCode: http.StatusBadRequest},
			body: `{"log": "test", "pass": "test"}`,
		},
		{
			name: "Empty request",
			want: want{statusCode: http.StatusBadRequest},
			body: "",
		},
		{
			name: "Empty JSON request",
			want: want{
				statusCode: http.StatusBadRequest,
			},
			body: `{}`,
		},
		{
			name: "Empty",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl, repo, uc := controller(t)
			repo.EXPECT().Do().Return(uc).Times(2)
			switch {
			case tt.name == "Internal server error":
				repo.EXPECT().DoRegister(context.Background(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("internal server error")).Times(1)
			case tt.name == "Duplicate login":
				// создаем экземпляр pq.Error
				err := pq.Error{
					Message: "pq: duplicate key value violates unique constraint \"users_pkey\"",
					Code:    "23505",
				}
				repo.EXPECT().DoRegister(context.Background(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&err).Times(1)
			default:
				repo.EXPECT().DoRegister(context.Background(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			}
			r, err := http.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()
			handler := http.HandlerFunc(ctrl.Register)
			handler(w, r)
			require.NoError(t, err)
			assert.Equal(t, w.Code, tt.want.statusCode, "Код ответа не совпадает с ожидаемым")
		})
	}
}

func TestAuthenticationHandler(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name string
		want want
		body string
	}{
		{
			name: "Valid auth",
			want: want{statusCode: http.StatusOK},
			body: `{"login": "test", "password": "12345"}`,
		},
		{
			name: "NoValid auth",
			want: want{statusCode: http.StatusUnauthorized},
			body: `{"login": "wrong", "password": ""}`,
		},
		{
			name: "Wrong request",
			want: want{statusCode: http.StatusBadRequest},
			body: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl, repo, uc := controller(t)
			repo.EXPECT().Do().Return(uc).Times(2)
			if tt.name == "NoValid auth" {
				repo.EXPECT().DoAuth(context.Background(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("unauthorized")).Times(1)
			}
			repo.EXPECT().DoAuth(context.Background(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			r, err := http.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()
			handler := http.HandlerFunc(ctrl.Authentication)
			handler(w, r)
			require.NoError(t, err)
			assert.Equal(t, w.Code, tt.want.statusCode, "Код ответа не совпадает с ожидаемым")
		})
	}
}
