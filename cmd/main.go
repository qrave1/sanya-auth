package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"os"
	"sanya-auth/internal/config"
	"sanya-auth/internal/domain"
	"sanya-auth/internal/repository"
	"time"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

var secret []byte

var repo *repository.GormRepository

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if creds.Username == "" || creds.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user := domain.User{Username: creds.Username, Password: string(hashedPassword)}
	ctx := context.Background()
	err = repo.CreateUser(ctx, &user)
	if err != nil {
		w.WriteHeader(http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	user, err := repo.GetUserByUsername(ctx, creds.Username)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)) != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(time.Hour * 24)
	claims := &Claims{
		Username: creds.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tokenRecord := domain.Token{UserID: user.ID, Token: tokenString, Expiration: expirationTime}
	err = repo.SetToken(ctx, &tokenRecord)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(tokenRecord)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func ValidateTokenHandler(w http.ResponseWriter, r *http.Request) {
	t := r.Header.Get("Authorization")
	if t == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(t, claims, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			log.Println(err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if !token.Valid {
		log.Println("token is not valid")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	ctx := context.Background()
	user, err := repo.GetUserByUsername(ctx, claims.Username)
	if err != nil {
		log.Println("error get user by username")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tokenRecord, err := repo.GetTokenByUserID(ctx, user.ID)
	if err != nil || t != tokenRecord.Token {
		log.Println("error get token by user id")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Logger - middleware, который логирует входящие запросы
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf(
			"%s\t%s\t%s",
			r.Method,
			r.RequestURI,

			time.Since(start),
		)
	})
}

func main() {
	cfg := config.NewConfig()

	secret = []byte(cfg.Server.Secret)

	log.SetOutput(os.Stdout)

	repos, err := repository.NewGormRepository(cfg)
	if err != nil {
		panic(err)
	}

	repo = repos

	http.Handle("/register", Logger(http.HandlerFunc(RegisterHandler)))
	http.Handle("/login", Logger(http.HandlerFunc(LoginHandler)))
	http.Handle("/validate", Logger(http.HandlerFunc(ValidateTokenHandler)))

	log.Println("Server is starting...")
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", cfg.Server.Port), nil))
}
