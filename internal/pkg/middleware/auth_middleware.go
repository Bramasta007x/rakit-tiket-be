package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"rakit-tiket-be/pkg/util"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type AuthMiddleware interface {
	VerifyToken(next echo.HandlerFunc) echo.HandlerFunc
	RequireAdmin(next echo.HandlerFunc) echo.HandlerFunc
}

type authMiddleware struct {
	log util.LogUtil
}

func MakeAuthMiddleware(log util.LogUtil) AuthMiddleware {
	return authMiddleware{
		log: log,
	}
}

// VerifyToken: Validasi JWT Token (Apakah user login?)
func (m authMiddleware) VerifyToken(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Missing Authorization Header")
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid Token Format")
		}

		// Parse Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			// Pastikan secret key SAMA PERSIS dengan yang di auth_service.go
			return []byte(util.BuildJwtSecret("rakit-tiket-app")), nil
		})

		if err != nil || !token.Valid {
			m.log.Error(c.Request().Context(), "Token Invalid: "+err.Error())
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or Expired Token")
		}

		// Simpan claims ke context agar bisa dibaca handler lain/middleware selanjutnya
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("user_id", claims["sub"])
			c.Set("role", claims["role"])
		}

		return next(c)
	}
}

// RequireAdmin: Validasi Role (Apakah user adalah ADMIN?)
func (m authMiddleware) RequireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Pastikan VerifyToken sudah dijalankan sebelumnya
		role, ok := c.Get("role").(string)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "User role not found")
		}

		if role != "ADMIN" {
			return echo.NewHTTPError(http.StatusForbidden, "Access Denied: Admins Only")
		}

		return next(c)
	}
}
