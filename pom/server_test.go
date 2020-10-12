package pom_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"pomerium-interview-rs-wilson/auth"
	"pomerium-interview-rs-wilson/pom"
	"pomerium-interview-rs-wilson/pom/mocks"
	"pomerium-interview-rs-wilson/stats"

	"github.com/gorilla/mux"
)

type sessionCreateTestCase struct {
	name                 string
	expectedResponseCode int
	mockStore            mocks.MockConfig
}

func TestServer_SessionCreate(t *testing.T) {
	// make test cases
	testCases := []sessionCreateTestCase{
		sessionCreateTestCase{
			name:                 "Success",
			expectedResponseCode: 200,
			mockStore: mocks.MockConfig{
				UserCheck: true,
				PassCheck: true,
			},
		},
		sessionCreateTestCase{
			name:                 "invalid user",
			expectedResponseCode: 404,
			mockStore: mocks.MockConfig{
				UserCheck: false,
				PassCheck: true,
			},
		},
		sessionCreateTestCase{
			name:                 "invalid password",
			expectedResponseCode: 403,
			mockStore: mocks.MockConfig{
				UserCheck: true,
				PassCheck: false,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// setup
			server := pom.NewServer(":7357", &testCase.mockStore, &testCase.mockStore, stats.NewKeeper([]string{"bob"}), auth.NewJWTAuth("test-secret-key"))

			res := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "http://localhost/session/create/", nil)
			if err != nil {
				t.Errorf("failed to create http request: %s", err.Error())
				return
			}
			req.Header.Set("Authorization", "Basic Ym9iOiQyYSQxMCRKMFdYcnZyamZqZ1hHUm9qV1VZWDd1U3pxMzVId1FxWVZqN2RqdEEwWUh1dDAua2FDY1VSZQ==")

			// test
			server.HandleSessionCreate(res, req)

			// verify
			if res.Result().StatusCode != testCase.expectedResponseCode {
				t.Errorf("Got status code %d: expected %d", res.Result().StatusCode, testCase.expectedResponseCode)
			}
		})
	}
}

type statsTestCase struct {
	name                 string
	token                string
	expectedResponseCode int
}

func TestServer_StatsRequest(t *testing.T) {
	//Setup auth
	realAuth := auth.NewJWTAuth("test-secret-key")
	bobToken, err := realAuth.GetToken("bob")
	eveToken, err := realAuth.GetToken("eve")
	if err != nil {
		t.Errorf("got error from auth: %s", err.Error())
		return
	}

	// make test cases
	testCases := []statsTestCase{
		statsTestCase{
			name:                 "Success",
			token:                bobToken,
			expectedResponseCode: 200,
		},
		statsTestCase{
			name:                 "invalid token",
			token:                "not really a thing",
			expectedResponseCode: 401,
		},
		statsTestCase{
			name:                 "forbidden user",
			token:                eveToken,
			expectedResponseCode: 403,
		},
	}

	// Run test cases
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// setup
			server := pom.NewServer(":7357", &mocks.MockConfig{}, &mocks.MockConfig{}, stats.NewKeeper([]string{"bob"}), realAuth)

			res := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "http://localhost/session/bob/stats", nil)
			if err != nil {
				t.Error(err)
			}
			req = mux.SetURLVars(req, map[string]string{"user": "bob"}) //inject testable context

			//TODO: test middleware separately?
			token := fmt.Sprintf("Bearer %s", testCase.token)
			req.Header.Set("Authorization", token)

			// test
			server.AuthMiddleware(server.HandleStatsRequest)(res, req)

			// verify
			if res.Result().StatusCode != testCase.expectedResponseCode {
				t.Errorf("Got status code %d: expected %d", res.Result().StatusCode, testCase.expectedResponseCode)
				return
			}
		})
	}
}
