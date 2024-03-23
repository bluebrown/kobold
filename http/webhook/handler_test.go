package webhook

import (
	"context"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestWebhook_ServeHTTP(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		giveUrl     string
		giveBody    string
		wantStatus  int
		wantResBody string
		wantChan    string
	}{
		{
			name:        "query parameter",
			giveUrl:     "/events?chan=query",
			giveBody:    "hello",
			wantStatus:  200,
			wantResBody: warnQueryDeprecated,
			wantChan:    "query",
		},
		{
			name:       "path parameter",
			giveUrl:    "/events/path",
			giveBody:   "hello",
			wantStatus: 202,
			wantChan:   "path",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var (
				ms  = mockScheduler{}
				api = New(&ms)
				rec = httptest.NewRecorder()
				req = httptest.NewRequest("POST", tt.giveUrl, strings.NewReader(tt.giveBody))
			)
			api.ServeHTTP(rec, req)
			res := rec.Result()
			assertEq(t, res.StatusCode, tt.wantStatus)
			assertEq(t, rec.Body.String(), tt.wantResBody)
			assertEq(t, ms.ch, tt.wantChan)
			assertEq(t, string(ms.buf), tt.giveBody)
		})
	}
}

type mockScheduler struct {
	ch  string
	buf []byte
}

func (m *mockScheduler) Schedule(ctx context.Context, chn string, data []byte) error {
	m.ch = chn
	m.buf = data
	return nil
}

func assertEq[T any](t *testing.T, got, want T) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
