package auth

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const tokenDuration = 3600 * time.Second
const invalJtiCleanRate = 1 * time.Hour

var ListenerChan chan bool

var state struct {
	sync.RWMutex
	InvalJti map[string]*jwt.NumericDate
	NTokens  int
}

type User struct {
	Login string `json:"login" binding:"required"`
	Role  string `json:"role" binding:"required"`
}

type UserClaims struct {
	User
	jwt.RegisteredClaims
}

func secret() string {
	return os.Getenv("SECRET")
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
				return []byte(secret()), nil
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

		state.RLock()
		if _, ok := state.InvalJti[claims.ID]; ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "expired JWT",
			})
			state.RUnlock()
			return
		}

		state.RUnlock()
		c.Next()
	}
}

func GetToken(user User) (string, error) {
	state.Lock()
	defer state.Unlock()

	claims := UserClaims{
		user,
		getJWTBasicTemplate(user.Login,
			fmt.Sprintf("%s-%d", user.Login, state.NTokens)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ts, err := token.SignedString([]byte(secret()))
	state.NTokens++
	return ts, err
}

func InvalidateToken(validTokenStr string) error {
	token, err := jwt.ParseWithClaims(validTokenStr, &UserClaims{},
		func(token *jwt.Token) (any, error) {
			return []byte(secret()), nil
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

	state.Lock()
	defer state.Unlock()

	exp, err := token.Claims.GetExpirationTime()
	if err != nil {
		return err
	}
	state.InvalJti[claims.ID] = exp
	return nil
}

func InvalJtiCleaner() {
	for {
		state.Lock()

		now := time.Now()
		for jti, exp := range state.InvalJti {
			if exp.Time.Before(now) {
				delete(state.InvalJti, jti)
			}
		}

		state.Unlock()
		time.Sleep(invalJtiCleanRate)
	}
}

func init() {
	state.Lock()
	defer state.Unlock()

	state.InvalJti = make(map[string]*jwt.NumericDate, 0)
	state.NTokens = 0

	go InvalJtiCleaner()
}
