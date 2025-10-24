package auth

import (
    "time"

    "github.com/golang-jwt/jwt/v5"
)

var jwtTTL = 24 * time.Hour // token expires in 1 day

func GenerateJWT(userID, role string, tenantID *string, secret string) (string, error) {
    claims := jwt.MapClaims{
        "sub":  userID,
        "role": role,
        "exp":  time.Now().Add(jwtTTL).Unix(),
    }

    if tenantID != nil {
        claims["tenant_id"] = *tenantID
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}
