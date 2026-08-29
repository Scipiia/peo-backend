package auth_ldap

import "github.com/golang-jwt/jwt/v5"

type User struct {
	UID        string
	FullName   string
	Permission string
	DN         string
	Email      string
}

type CustomClaims struct {
	UID         string   `json:"uid"`
	FullName    string   `json:"full_name"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}
