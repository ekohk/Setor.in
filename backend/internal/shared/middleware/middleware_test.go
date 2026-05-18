package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/setorin/setorin/backend/internal/infrastructure/keycloak"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ─── Stub Validator ──────────────────────────────────────────────────────────

// stubValidator lets tests control what Validate returns without needing a real Keycloak.
type stubValidator struct {
	claims *keycloak.Claims
	err    error
}

func (s *stubValidator) Validate(_ string) (*keycloak.Claims, error) {
	return s.claims, s.err
}

// JWT accepts a validatorFunc so we can inject stubs in tests.
// We rewrite the middleware inline here to avoid changing production code.
func jwtWithStub(v *stubValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := extractBearer(c.GetHeader("Authorization"))
		claims, err := v.Validate(raw)
		if err != nil {
			respondAuthError(c, err)
			return
		}
		c.Set(ctxKeyClaims, claims)
		c.Next()
	}
}

func claimsWithRoles(roles ...string) *keycloak.Claims {
	return &keycloak.Claims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: "test-sub"},
		RealmAccess:      keycloak.RealmAccess{Roles: roles},
	}
}

func perform(r *gin.Engine, method, path string, authHeader string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	r.ServeHTTP(w, req)
	return w
}

func bodyCode(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	var env struct {
		Error *struct{ Code string } `json:"error"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&env))
	if env.Error == nil {
		return ""
	}
	return env.Error.Code
}

// ─── JWT middleware ──────────────────────────────────────────────────────────

func TestJWT_MissingToken_Returns401(t *testing.T) {
	r := gin.New()
	stub := &stubValidator{err: keycloak.ErrTokenMissing}
	r.GET("/protected", jwtWithStub(stub), func(c *gin.Context) { c.Status(200) })

	w := perform(r, "GET", "/protected", "")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "UNAUTHORIZED", bodyCode(t, w))
}

func TestJWT_InvalidToken_Returns401(t *testing.T) {
	r := gin.New()
	stub := &stubValidator{err: keycloak.ErrTokenInvalid}
	r.GET("/protected", jwtWithStub(stub), func(c *gin.Context) { c.Status(200) })

	w := perform(r, "GET", "/protected", "Bearer invalid.token.here")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "TOKEN_INVALID", bodyCode(t, w))
}

func TestJWT_ExpiredToken_Returns401(t *testing.T) {
	r := gin.New()
	stub := &stubValidator{err: keycloak.ErrTokenExpired}
	r.GET("/protected", jwtWithStub(stub), func(c *gin.Context) { c.Status(200) })

	w := perform(r, "GET", "/protected", "Bearer expired.token.here")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "TOKEN_EXPIRED", bodyCode(t, w))
}

func TestJWT_ValidToken_Passes(t *testing.T) {
	r := gin.New()
	stub := &stubValidator{claims: claimsWithRoles("user")}
	var gotClaims *keycloak.Claims
	r.GET("/protected", jwtWithStub(stub), func(c *gin.Context) {
		gotClaims = MustClaims(c)
		c.Status(200)
	})

	w := perform(r, "GET", "/protected", "Bearer valid.token.here")

	assert.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, gotClaims)
	assert.Equal(t, "test-sub", gotClaims.Subject)
}

func TestJWT_MalformedBearerPrefix_Returns401(t *testing.T) {
	r := gin.New()
	stub := &stubValidator{err: keycloak.ErrTokenMissing}
	r.GET("/protected", jwtWithStub(stub), func(c *gin.Context) { c.Status(200) })

	// "Token xyz" instead of "Bearer xyz"
	w := perform(r, "GET", "/protected", "Token somevalue")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ─── RequireRole middleware ──────────────────────────────────────────────────

func setupRBAC(roles ...string) (*gin.Engine, *stubValidator) {
	r := gin.New()
	stub := &stubValidator{claims: claimsWithRoles(roles...)}
	r.GET("/admin",
		jwtWithStub(stub),
		RequireRole(RoleAdmin),
		func(c *gin.Context) { c.Status(200) },
	)
	r.GET("/collector",
		jwtWithStub(stub),
		RequireRole(RoleCollector),
		func(c *gin.Context) { c.Status(200) },
	)
	r.GET("/multi",
		jwtWithStub(stub),
		RequireRole(RoleAdmin, RoleCollector),
		func(c *gin.Context) { c.Status(200) },
	)
	return r, stub
}

func TestRequireRole_CorrectRole_Returns200(t *testing.T) {
	r, _ := setupRBAC(RoleAdmin)
	w := perform(r, "GET", "/admin", "Bearer tok")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_WrongRole_Returns403(t *testing.T) {
	r, _ := setupRBAC(RoleUser)
	w := perform(r, "GET", "/admin", "Bearer tok")
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, "INSUFFICIENT_ROLE", bodyCode(t, w))
}

func TestRequireRole_MultipleAllowed_MatchFirst(t *testing.T) {
	r, _ := setupRBAC(RoleAdmin)
	w := perform(r, "GET", "/multi", "Bearer tok")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_MultipleAllowed_MatchSecond(t *testing.T) {
	r, _ := setupRBAC(RoleCollector)
	w := perform(r, "GET", "/multi", "Bearer tok")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_SuperAdminWithoutAdminRole_Returns403(t *testing.T) {
	// RequireRole does a plain membership check — super_admin is NOT an implicit
	// bypass for other roles. Routes must explicitly include RoleSuperAdmin if needed.
	r, _ := setupRBAC(RoleSuperAdmin)
	w := perform(r, "GET", "/admin", "Bearer tok")
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireRole_SuperAdminExplicitlyListed_Returns200(t *testing.T) {
	r := gin.New()
	stub := &stubValidator{claims: claimsWithRoles(RoleSuperAdmin)}
	r.GET("/superroute",
		jwtWithStub(stub),
		RequireRole(RoleAdmin, RoleSuperAdmin), // both listed
		func(c *gin.Context) { c.Status(200) },
	)
	w := perform(r, "GET", "/superroute", "Bearer tok")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_EmptyRoles_Panics(t *testing.T) {
	assert.Panics(t, func() {
		RequireRole() // must panic: misuse guard
	})
}

// ─── extractBearer helper ────────────────────────────────────────────────────

func TestExtractBearer(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"Bearer mytoken", "mytoken"},
		{"bearer mytoken", "mytoken"}, // case-insensitive
		{"BEARER mytoken", "mytoken"},
		{"Bearer  mytoken", "mytoken"}, // extra space trimmed
		{"", ""},
		{"Token mytoken", ""},  // wrong prefix
		{"Bearer", ""},         // no token after prefix
		{"mytoken", ""},        // no prefix at all
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, extractBearer(tc.in), "input: %q", tc.in)
	}
}

// ─── MustClaims ─────────────────────────────────────────────────────────────

func TestMustClaims_PanicsWhenMissing(t *testing.T) {
	r := gin.New()
	r.GET("/oops", func(c *gin.Context) {
		// MustClaims without JWT middleware set — should panic
		assert.Panics(t, func() { MustClaims(c) })
		c.Status(200)
	})
	perform(r, "GET", "/oops", "")
}

func TestClaims_ReturnsFalseWhenMissing(t *testing.T) {
	r := gin.New()
	r.GET("/ok", func(c *gin.Context) {
		cl, ok := Claims(c)
		assert.False(t, ok)
		assert.Nil(t, cl)
		c.Status(200)
	})
	perform(r, "GET", "/ok", "")
}
