package security

import (
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	COOKIE_REFRESH_TOKEN = "__Cat_Garong"
	COOKIE_ACCESS_TOKEN  = "__Cat_Baik"
)

func SetRefreshToken(w http.ResponseWriter, d time.Duration, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     COOKIE_REFRESH_TOKEN,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteNoneMode,
		Expires:  time.Now().Add(d),
	})
}

func DeleteRefreshToken(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     COOKIE_REFRESH_TOKEN,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}

func SetAccessToken(w http.ResponseWriter, d time.Duration, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     COOKIE_ACCESS_TOKEN,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteNoneMode,
		Expires:  time.Now().Add(d),
	})
}

func DeleteAccessToken(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     COOKIE_ACCESS_TOKEN,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}

func GetRefreshToken(r *http.Request) (token string, id bson.ObjectID) {
	cookie, err := r.Cookie(COOKIE_REFRESH_TOKEN)
	if err != nil {
		return "", bson.NilObjectID
	}

	splitValue := strings.Split(cookie.Value, ".")
	if len(splitValue) != 2 {
		return "", bson.NilObjectID
	}

	objectId, err := bson.ObjectIDFromHex(strings.ToLower(splitValue[1]))
	if err != nil {
		return "", bson.NilObjectID
	}

	return splitValue[0], objectId
}

func GetAccessToken(r *http.Request) string {
	cookie, err := r.Cookie(COOKIE_ACCESS_TOKEN)
	if err != nil {
		return ""
	}
	return cookie.Value
}
