package api

import (
	mockdb "BankAppGo/db/mock"
	db "BankAppGo/db/sqlc"
	"BankAppGo/db/util"
	"BankAppGo/token"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type eqCreateAccountParamsMatcher struct {
	arg   db.CreateAccountParams
	owner string
}

// checks if two accounts are the same
func (e eqCreateAccountParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateAccountParams)
	if !ok {
		return false
	}
	err := util.IsSupportedCurrency(arg.Currency)
	if !err {
		return false
	}
	return reflect.DeepEqual(e.arg, arg)
}

func (e eqCreateAccountParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and owner %v", e.arg, e.owner)
}

func EqCreateAccountParams(arg db.CreateAccountParams, owner string) gomock.Matcher {
	return eqCreateAccountParamsMatcher{arg, owner}
}

func TestGetAccountAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	testCases := []struct {
		name          string
		accountID     int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//check response from http request
				require.Equal(t, http.StatusOK, recorder.Code)

				//check returned account from API
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "UnauthorizedUser",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user.Username", time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//check response from http request
				require.Equal(t, http.StatusUnauthorized, recorder.Code)

			},
		},
		{
			name:      "NoAuthorization",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//check response from http request
				require.Equal(t, http.StatusUnauthorized, recorder.Code)

			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//check response from http request
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//check response from http request
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			accountID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(0)).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//check response from http request
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			// builds stub
			tc.buildStubs(store)
			//start test server and send request

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			// check response
			tc.checkResponse(t, recorder)

		})
	}

}

func TestCreateAccount(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	testCases := []struct {
		name          string
		account       db.Account
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:    "OK",
			account: account,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateAccountParams{
					Owner:    user.Username,
					Currency: account.Currency,
					Balance:  0,
				}
				store.EXPECT().CreateAccount(gomock.Any(), EqCreateAccountParams(arg, user.Username)).Times(1).Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//check response from http request
				require.Equal(t, http.StatusOK, recorder.Code)

				//check returned account from API
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},

		{
			name:    "UnauthorizedUser",
			account: account,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker) {
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//check response from http request
				require.Equal(t, http.StatusUnauthorized, recorder.Code)

			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			// builds stub
			tc.buildStubs(store)
			//start test server and send request

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(map[string]string{
				"currency": account.Currency,
			})
			require.NoError(t, err)
			url := "/accounts"

			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			// check response
			tc.checkResponse(t, recorder)

		})
	}

}

func TestListAccounts(t *testing.T) {
	user, _ := randomUser(t)
	n := 5
	var accounts []db.Account
	for i := 0; i < n; i++ {
		accounts = append(accounts, randomAccount(user.Username))
	}

	testCases := []struct {
		name          string
		query         gin.H
		accounts      []db.Account
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: gin.H{
				"page_size": n,
				"page_id":   1,
			},
			accounts: accounts,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker) {
				// adding token authorization (either JWT or PASETO)
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListAccountsParams{
					Owner:  user.Username,
					Limit:  int64(n),
					Offset: 0,
				}

				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(accounts, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//check response from http request
				require.Equal(t, http.StatusOK, recorder.Code)

				//check returned account from API
				requireBodyMatchAccountList(t, recorder.Body, accounts)
			},
		},
		{
			name: "InvalidPageID",
			query: gin.H{
				"page_size": n,
				"page_id":   0,
			},
			accounts: accounts,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker) {
				// adding token authorization (either JWT or PASETO)
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListAccountsParams{
					Owner:  user.Username,
					Limit:  int64(n),
					Offset: -1,
				}

				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//check response from http request
				require.Equal(t, http.StatusBadRequest, recorder.Code)

			},
		},
		{
			name: "InvalidPageSize",
			query: gin.H{
				"page_size": 100,
				"page_id":   1,
			},
			accounts: accounts,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker) {
				// adding token authorization (either JWT or PASETO)
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListAccountsParams{
					Owner:  user.Username,
					Limit:  10,
					Offset: 0,
				}

				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//check response from http request
				require.Equal(t, http.StatusBadRequest, recorder.Code)

			},
		},
		{
			name: "Unauthorized",
			query: gin.H{
				"page_size": n,
				"page_id":   1,
			},
			accounts: accounts,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListAccountsParams{
					Owner:  "",
					Limit:  int64(n),
					Offset: 0,
				}

				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//check response from http request
				require.Equal(t, http.StatusUnauthorized, recorder.Code)

			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			// builds stub
			tc.buildStubs(store)
			//start test server and send request

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			data := GinHToURLValues(tc.query)
			url := "/accounts?" + data.Encode()

			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			// check response
			tc.checkResponse(t, recorder)

		})
	}

}

func randomAccount(owner string) db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    owner,
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}

func requireBodyMatchAccountList(t *testing.T, body *bytes.Buffer, accounts []db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccounts []db.Account
	err = json.Unmarshal(data, &gotAccounts)
	require.NoError(t, err)

	require.Equal(t, accounts, gotAccounts)

}
