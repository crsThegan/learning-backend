package auth

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
	"time"
)

const secret = "secret-key"
const tokenDuration time.Duration = 3600 * time.Second

var invalJti map[string]bool
var generatedTokens int

type User struct {
	Login string `json:"login" binding:"required"`
	Role  string `json:"role" binding:"required"`
}

type UserClaims struct {
	User
	jwt.RegisteredClaims
}

func getJWTBasicTemplate(sub, jti string) jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		Subject:   sub,
		Issuer:    "localhost",
		Audience:  []string{"front", "postman"},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenDuration)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ID:        jti,
	}
}

func Require(role string) func(c *gin.Context) {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "missing bearer token",
			})
			return
		}

		receivedStr := strings.TrimPrefix(h, "Bearer ")
		received, err := jwt.ParseWithClaims(receivedStr, &UserClaims{},
			func(token *jwt.Token) (any, error) {
				return []byte(secret), nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil || !received.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "not a valid JWT",
			})
			return
		}

		claims, ok := received.Claims.(*UserClaims)
		if !ok || (role != "any" && claims.Role != role) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "incorrect JWT claims",
			})
			return
		}

		if _, ok := invalJti[claims.ID]; ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "expired JWT",
			})
		}
		c.Next()
	}
}

func GetToken(user User) (string, error) {
	claims := UserClaims{
		user,
		getJWTBasicTemplate(user.Login,
			fmt.Sprintf("%s-%d", user.Login, generatedTokens)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ts, err := token.SignedString([]byte(secret))
	generatedTokens++
	return ts, err
}

func InvalidateToken(validTokenStr string) error {
	token, err := jwt.ParseWithClaims(validTokenStr, &UserClaims{},
		func(token *jwt.Token) (any, error) {
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return errors.New("invalid token: " + err.Error())
	} else if !token.Valid {
		return errors.New("invalid token")
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return errors.New("invalid token claims format")
	}

	invalJti[claims.ID] = true
	return nil
}

func init() {
	invalJti = make(map[string]bool, 0)
	generatedTokens = 0
}
