package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"q-dev/http/common"
	"q-dev/infra/mysql/dao"
	"q-dev/pkg/testutil"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupMockDB(t *testing.T) sqlmock.Sqlmock {
	t.Helper()
	gormDB, mock, err := testutil.NewMockDB()
	require.NoError(t, err)
	dao.SetDefault(gormDB)
	return mock
}

func TestHelloWorldAPI_List_Success(t *testing.T) {
	mock := setupMockDB(t)

	// FindByPage first does SELECT * with ORDER BY, LIMIT, OFFSET
	// When result size (1) < limit (10), it skips the count query
	// and computes count = size + offset = 1 + 0 = 1
	mock.ExpectQuery("SELECT \\*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "created_at", "updated_at"}).
			AddRow(1, "test", "desc", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/hello-world?page=1&page_size=10", nil)

	h := NewHelloWorldAPI()
	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp common.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHelloWorldAPI_List_MissingParams(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/hello-world", nil) // missing page, page_size

	h := NewHelloWorldAPI()
	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp common.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, -1, resp.Code) // param validation failure
}

func TestHelloWorldAPI_Get_Success(t *testing.T) {
	mock := setupMockDB(t)

	mock.ExpectQuery("SELECT \\*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "created_at", "updated_at"}).
			AddRow(1, "test", "desc", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/hello-world/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	h := NewHelloWorldAPI()
	h.Get(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp common.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHelloWorldAPI_Get_InvalidID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/hello-world/abc", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	h := NewHelloWorldAPI()
	h.Get(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp common.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, -1, resp.Code) // parse ID failure
}
