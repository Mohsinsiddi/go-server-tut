package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/Mohsinsiddi/simplebank/db/mock"
	db "github.com/Mohsinsiddi/simplebank/db/sqlc"
	"github.com/Mohsinsiddi/simplebank/util"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestGetAccountAPI(t *testing.T){
    account := randomAccount()

	testCases := []struct{
		name string
		accoundtID int64
		buildStubs func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:"Ok",
			accoundtID:account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				// build stubs
					store.EXPECT().
					GetAccount(gomock.Any(),gomock.Eq(account.ID)).
					Times(1).
					Return(account,nil)	
			},
			checkResponse :func(t *testing.T, recorder *httptest.ResponseRecorder){
				// check response
					require.Equal(t, http.StatusOK, recorder.Code)
					requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:"NotFound",
			accoundtID:account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				// build stubs
					store.EXPECT().
					GetAccount(gomock.Any(),gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{},sql.ErrNoRows)	
			},
			checkResponse :func(t *testing.T, recorder *httptest.ResponseRecorder){
				// check response
					require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:"InternalError",
			accoundtID:account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				// build stubs
					store.EXPECT().
					GetAccount(gomock.Any(),gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{},sql.ErrConnDone)	
			},
			checkResponse :func(t *testing.T, recorder *httptest.ResponseRecorder){
				// check response
					require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:"InvalidID",
			accoundtID:0,
			buildStubs: func(store *mockdb.MockStore) {
				// build stubs
					store.EXPECT().
					GetAccount(gomock.Any(),gomock.Any()).
					Times(0)
			},
			checkResponse :func(t *testing.T, recorder *httptest.ResponseRecorder){
				// check response
					require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}
	
	for i:= range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			// start the test server and send request

			server := newTestServer(t,store)

			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d",tc.accoundtID) 
			request ,err := http.NewRequest(http.MethodGet, url, nil)	
			
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t,recorder)
		})
		
	}

	
	
}

func randomAccount() db.Account {
	return db.Account{
		ID: util.RandomInt(1,1000),
        Owner: util.RandomOwner(),
		Balance: util.RandomMoney(),
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