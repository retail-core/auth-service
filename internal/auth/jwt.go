package auth

import (
    "time"

    "github.com/golang-jwt/jwt/v5"
)

var jwtTTL = 24 * time.Hour // token expires in 1 day

func GenerateJWT(userID, role string, tenantID *string, isVerified bool, secret string) (string, error) {
    claims := jwt.MapClaims{
        "sub":  userID,
        "role": role,
        "exp":  time.Now().Add(jwtTTL).Unix(),
        "is_verified": isVerified,
        "iss": "auth-issuer",
    }

    if tenantID != nil {
        claims["tenant_id"] = *tenantID
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}

func GenerateRefreshToken(userID, secret string) (string, time.Time, error) {
    expiresAt := time.Now().Add(30 * 24 * time.Hour)
    claims := jwt.MapClaims{
        "sub": userID,
        "exp": expiresAt.Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    signedToken, err := token.SignedString([]byte(secret))
    if err != nil {
        return "", time.Time{}, err
    }
    return signedToken, expiresAt, nil
}