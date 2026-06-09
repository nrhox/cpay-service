package config

import (
	"encoding/base64"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/nrhox/cpay-service/internal/entity"
	"github.com/nrhox/cpay-service/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ModeApp int

const (
	MODE_DEBUG ModeApp = iota
	MODE_PRODUCTION
)

type Config struct {
	AppPort        string
	FrontendUrl    string
	Mode           ModeApp
	Mongo          Mongodb
	Session        Session
	Providers      Providers
	UserMock       UserMock
	SnowFlakeEpoch int
	MaxPaymentTIme time.Duration
	AllowOrigin    []string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system env")
	}

	return &Config{
		AppPort:        ":" + getEnv("PORT", "8080"),
		Mode:           parseAppMode(getEnv("MODE", "DEBUG")),
		FrontendUrl:    getEnv("FRONTEND_URL", "http://localhost:3003"),
		SnowFlakeEpoch: getIntEnv("SNOW_FLAKE_EPOCH", 1772298000000),
		MaxPaymentTIme: getDurationEnv("MAX_PAYMENT_TIME", 1*time.Hour),
		AllowOrigin:    strings.Split(getEnv("ALLOW_ORIGIN", "http://localhost:3000,http://localhost:5173,http://127.0.0.1:3000"), ","),
		Mongo: Mongodb{
			DbUrl:        getEnv("MONGODB_URI", "mongodb://localhost:27017"),
			DatabaseName: getEnv("MONGODB_DATABASE", "locker_app"),
		},
		Session: Session{
			RefreshDuration:     getDurationEnv("SESSION_REFRESH_DURATION", 30*24*time.Hour),
			AccessTokenDuration: getDurationEnv("SESSION_ACCESS_DURATION", 5*time.Minute),
			JwtPublicKey:        getSecret("JWT_PUBLIC_KEY"),
			JwtPrivateKey:       getSecret("JWT_PRIVATE_KEY"),
			SaltKey:             os.Getenv("SALT_KEY"),
		},
		Providers: Providers{
			Google: &OauthConfig{
				ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
				ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
				RedirectURL:  getEnv("GOOGLE_REDIRECT_URL", resolveUrlLocal(getEnv("PORT", "8080"), "/api/auth/google/callback")),
			},
			Github: &OauthConfig{
				ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
				ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
				RedirectURL:  getEnv("GITHUB_REDIRECT_URL", resolveUrlLocal(getEnv("PORT", "8080"), "/api/auth/github/callback")),
			},
		},
		UserMock: LoadUserMock(),
	}
}

func resolveUrlLocal(port string, path string) string {
	return "http://localhost:" + port + path
}

func parseAppMode(mode string) ModeApp {
	mode = strings.ToLower(mode)
	if mode == "PRODUCTIOn" {
		return MODE_PRODUCTION
	}
	return MODE_DEBUG
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	return utils.ParseDuration(value)
}

func getSecret(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		panic(key + " not found in env")
	}

	bValue, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		panic(key + " error: " + err.Error())
	}

	return string(bValue)
}

func LoadUserMock() UserMock {
	useMock := os.Getenv("USE_MOCK_USER") == "true"
	fileMock := os.Getenv("FILE_MOCK_USER")

	if useMock && fileMock != "" {
		data, err := os.ReadFile(fileMock)
		if err != nil {
			panic(err)
		}

		var user entity.User
		err = bson.UnmarshalExtJSON(data, true, &user)
		if err != nil {
			panic(err)
		}

		return UserMock{
			User:     &user,
			WithMock: useMock,
			MockFile: fileMock,
		}
	}
	return UserMock{
		User:     nil,
		WithMock: useMock,
		MockFile: fileMock,
	}
}

func getIntEnv(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if val, err := strconv.Atoi(value); err == nil {
			return val
		}
	}
	return fallback
}
