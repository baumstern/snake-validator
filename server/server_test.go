package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewGame(t *testing.T) {

	initialSnake := &snake{
		X:    0,
		Y:    0,
		VelX: 1,
		VelY: 0,
	}

	for _, test := range []struct {
		name           string
		urlPath        string
		wantStatusCode int
		wantState      state
	}{
		{
			name:           "invalid map size",
			urlPath:        "/new?w=1&h=1",
			wantStatusCode: 400,
			wantState:      state{},
		},
		{
			name:           "0 width",
			urlPath:        "/new?w=0&h=1",
			wantStatusCode: 400,
			wantState:      state{},
		},
		{
			name:           "negative width",
			urlPath:        "/new?w=-1&h=1",
			wantStatusCode: 400,
			wantState:      state{},
		},
		{
			name:           "alphabet width",
			urlPath:        "/new?w=0&h=1",
			wantStatusCode: 400,
			wantState:      state{},
		},
		{
			name:           "valid request",
			urlPath:        "/new?w=2&h=2",
			wantStatusCode: 200,
			wantState: state{
				Width:  2,
				Height: 2,
				Snake:  *initialSnake,
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()

			h := AppHandler(NewGame)
			h.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.urlPath, nil))

			if response.Code != test.wantStatusCode {
				t.Errorf("got %d, want %d", response.Code, test.wantStatusCode)
			}

			if response.Code == 200 {
				got := state{}
				err := json.Unmarshal(response.Body.Bytes(), &got)
				if err != nil {
					t.Error("Failed to unmarshal response:", err)
				}
				// prevent randomized fruit position
				got.Fruit = fruit{}

				if !cmp.Equal(got, test.wantState) {
					t.Errorf("want %+v, got %+v", test.wantState, got)
				}
			}
		})
	}
}

func TestServer405(t *testing.T) {
	for _, test := range []struct {
		name           string
		method         string
		urlPath        string
		handler        func(http.ResponseWriter, *http.Request) *appError
		wantStatusCode int
	}{
		{
			name:           "(POST) invalid method to NewGame handler",
			method:         http.MethodPost,
			urlPath:        "/new?w=2&h=2",
			handler:        NewGame,
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "(PUT) invalid method to NewGame handler",
			method:         http.MethodPut,
			urlPath:        "/new?w=2&h=2",
			handler:        NewGame,
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "(DELETE) invalid method to NewGame handler",
			method:         http.MethodDelete,
			urlPath:        "/new?w=2&h=2",
			handler:        NewGame,
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "(GET) invalid method to Validate handler",
			method:         http.MethodGet,
			urlPath:        "/validate",
			handler:        Validate,
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "(PUT) invalid method to Validate handler",
			method:         http.MethodPut,
			urlPath:        "/validate",
			handler:        Validate,
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "(DELETE) invalid method to Validate handler",
			method:         http.MethodDelete,
			urlPath:        "/validate",
			handler:        Validate,
			wantStatusCode: http.StatusMethodNotAllowed,
		},
	} {
		t.Run("invalid method (POST)", func(t *testing.T) {
			request, _ := http.NewRequest(test.method, test.urlPath, nil)
			response := httptest.NewRecorder()

			h := AppHandler(test.handler)
			h.ServeHTTP(response, request)

			if response.Code != test.wantStatusCode {
				t.Errorf("%q: got status code = %d, want %d", test.urlPath, response.Code, test.wantStatusCode)
			}
		})
	}
}

func TestValidate(t *testing.T) {

	for _, test := range []struct {
		name           string
		urlPath        string
		body           payloadValidate
		wantStatusCode int
		wantState      state
	}{
		{
			name:    "catch fruit in 2x2",
			urlPath: "/validate",
			body: payloadValidate{
				GameID: "0",
				Width:  2,
				Height: 2,
				Score:  0,
				Fruit:  fruit{1, 0},
				Snake: snake{
					X:    0,
					Y:    0,
					VelX: 1,
					VelY: 0,
				},
				Ticks: []tick{{1, 0}},
			},
			wantStatusCode: 200,
			wantState: state{
				GameID: "0",
				Width:  2,
				Height: 2,
				Score:  1,
				Fruit:  fruit{},
				Snake: snake{
					X:    1,
					Y:    0,
					VelX: 1,
					VelY: 0,
				},
			},
		},
		{
			name:    "catch fruit in bottom 1x3",
			urlPath: "/validate",
			body: payloadValidate{
				GameID: "0",
				Width:  1,
				Height: 3,
				Score:  0,
				Fruit:  fruit{0, 2},
				Snake: snake{
					X:    0,
					Y:    0,
					VelX: 1,
					VelY: 0,
				},
				Ticks: []tick{{0, 1}, {0, 1}},
			},
			wantStatusCode: 200,
			wantState: state{
				GameID: "0",
				Width:  1,
				Height: 3,
				Score:  1,
				Fruit:  fruit{},
				Snake: snake{
					X:    0,
					Y:    2,
					VelX: 0,
					VelY: 1,
				},
			},
		},
		{
			name:    "invalid move diagonal",
			urlPath: "/validate",
			body: payloadValidate{
				GameID: "0",
				Width:  2,
				Height: 2,
				Score:  0,
				Fruit:  fruit{0, 1},
				Snake: snake{
					X:    0,
					Y:    0,
					VelX: 1,
					VelY: 0,
				},
				Ticks: []tick{{1, 1}},
			},
			wantStatusCode: 418,
			wantState:      state{},
		},
		{
			name:    "invalid move backward",
			urlPath: "/validate",
			body: payloadValidate{
				GameID: "0",
				Width:  2,
				Height: 2,
				Score:  0,
				Fruit:  fruit{1, 1},
				Snake: snake{
					X:    0,
					Y:    0,
					VelX: 1,
					VelY: 0,
				},
				Ticks: []tick{{1, 0}, {-1, 0}},
			},
			wantStatusCode: 418,
			wantState:      state{},
		},
		{
			name:    "out of bound",
			urlPath: "/validate",
			body: payloadValidate{
				GameID: "0",
				Width:  2,
				Height: 2,
				Score:  0,
				Fruit:  fruit{1, 1},
				Snake: snake{
					X:    0,
					Y:    0,
					VelX: 1,
					VelY: 0,
				},
				Ticks: []tick{{1, 0}, {1, 0}},
			},
			wantStatusCode: 418,
			wantState:      state{},
		},
		{
			name:    "fruit not found",
			urlPath: "/validate",
			body: payloadValidate{
				GameID: "0",
				Width:  2,
				Height: 2,
				Score:  0,
				Fruit:  fruit{1, 1},
				Snake: snake{
					X:    0,
					Y:    0,
					VelX: 1,
					VelY: 0,
				},
				Ticks: []tick{{1, 0}, {-1, 0}},
			},
			wantStatusCode: 418,
			wantState:      state{},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()

			h := AppHandler(Validate)
			body, _ := json.Marshal(test.body)
			h.ServeHTTP(response, httptest.NewRequest(http.MethodPost, test.urlPath, bytes.NewBuffer(body)))

			if response.Code != test.wantStatusCode {
				t.Errorf("got %d, want %d", response.Code, test.wantStatusCode)
			}

			if response.Code == 200 {
				got := payloadValidate{}
				err := json.Unmarshal(response.Body.Bytes(), &got)
				if err != nil {
					t.Error("Failed to unmarshal response:", err)
				}
				// prevent randomized fruit position
				got.Fruit = fruit{}
				state := state{
					GameID: got.GameID,
					Width:  got.Width,
					Height: got.Height,
					Score:  got.Score,
					Fruit:  got.Fruit,
					Snake:  got.Snake,
				}
				if !cmp.Equal(state, test.wantState) {
					t.Errorf("want %+v, got %+v", test.wantState, state)
				}
			}
		})
	}
}
