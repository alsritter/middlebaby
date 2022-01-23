package proxy

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestImposterHandler(t *testing.T) {
	bodyRequest := []byte(`
{
		"data": {
				"type": "gophers",
				"attributes": {
						"name": "Zebediah",
						"color": "Purple",
						"age": 55
					}
			}
}`)

	var headers = make(map[string]string)
	headers["Content-Type"] = "application/json"

	validRequest := Request{
		Method:  "POST",
		Headers: &headers,
	}

	body := `{"test":true}`

	tests := []struct {
		name         string
		imposter     Imposter
		expectedBody string
		statusCode   int
	}{
		{"valid imposter with body", Imposter{Request: validRequest, Response: Response{Status: http.StatusOK, Headers: &headers, Body: body}}, body, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req, err := http.NewRequest("POST", "/gophers", bytes.NewBuffer(bodyRequest))
			if err != nil {
				t.Fatalf("could not created request: %v", err)
			}

			rec := httptest.NewRecorder()
			handler := http.HandlerFunc(ImposterHandler(tt.imposter))
			handler.ServeHTTP(rec, req)

			if status := rec.Code; status != tt.statusCode {
				t.Errorf("handler expected %d code and got: %d code", tt.statusCode, status)
			}

			if rec.Body.String() != tt.expectedBody {
				t.Errorf("handler expected %s body and got: %s body", tt.expectedBody, rec.Body.String())
			}
		})
	}
}
