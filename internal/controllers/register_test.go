package controllers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nextlag/gomart/internal/config"
	"github.com/nextlag/gomart/internal/controllers/mocks"
	"github.com/nextlag/gomart/internal/mw/logger"
	"github.com/nextlag/gomart/internal/usecase"
)

func controller(t *testing.T) (*Controller, *mocks.MockUseCase, *usecase.UseCase) {
	t.Helper()
	log := logger.SetupLogger()
	var cfg config.HTTPServer
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	repo := mocks.NewMockUseCase(mockCtl)
	db := usecase.NewMockRepository(mockCtl)
	uc := usecase.New(db, log, cfg)

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
			want: want{
				statusCode: 200,
			},
			body: `{"login": "test", "password": "test"}`,
		},
		{
			name: "Empty password",
			want: want{
				statusCode: 200,
			},
			body: `{"login": "test"}`,
		},
		{
			name: "Empty login",
			want: want{
				statusCode: 400,
			},
			body: `{"password": "12345"}`,
		},
		{
			name: "Invalid request",
			want: want{
				statusCode: 400,
			},
			body: `{"log": "test", "pass": "test"}`,
		},
		{
			name: "Empty request",
			want: want{
				statusCode: 400,
			},
			body: "",
		},
		{
			name: "Empty JSON request",
			want: want{
				statusCode: 400,
			},
			body: `{}`,
		},
		{
			name: "Empty",
			want: want{
				statusCode: 400,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			ctrl, repo, uc := controller(t)
			repo.EXPECT().Do().Return(uc).Times(2)
			repo.EXPECT().DoRegister(context.Background(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			r, err := http.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()
			handler := http.HandlerFunc(ctrl.Register)
			handler(w, r)
			require.NoError(t, err)
			assert.Equal(t, w.Code, tt.want.statusCode, "Код ответа не совпадает с ожидаемым")
		})
	}
}
