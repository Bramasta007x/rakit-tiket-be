package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"rakit-tiket-be/internal/app/app_auth/dao"
	entity "rakit-tiket-be/pkg/entity/app_auth"
	"rakit-tiket-be/pkg/util"
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (string, error)
}

type authService struct {
	log   util.LogUtil
	sqlDB *sql.DB
}

func MakeAuthService(log util.LogUtil, sqlDB *sql.DB) AuthService {
	return authService{
		log:   log,
		sqlDB: sqlDB,
	}
}

func (s authService) Login(ctx context.Context, email, password string) (string, error) {
	dbTrx := dao.NewTransaction(ctx, s.sqlDB)
	// Tidak perlu defer Rollback karena hanya SELECT

	users, err := dbTrx.GetUserDAO().Search(ctx, entity.UserQuery{
		Emails: []string{email},
	})

	if err != nil {
		return "", err
	}
	if len(users) == 0 {
		return "", fmt.Errorf("invalid email or password")
	}

	user := users[0]

	// Verifikasi Password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", fmt.Errorf("invalid email or password")
	}

	// Generate JWT
	token, err := s.generateToken(user)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s authService) generateToken(user entity.UserEntity) (string, error) {
	// Menggunakan util.BuildJwtSecret yang sudah ada di codebase
	secretKey := util.BuildJwtSecret("rakit-tiket-app")

	claims := jwt.MapClaims{
		"sub":  user.ID,
		"name": user.Name,
		"role": user.Role,
		"exp":  time.Now().Add(time.Hour * 24).Unix(), // Token valid 24 jam
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
