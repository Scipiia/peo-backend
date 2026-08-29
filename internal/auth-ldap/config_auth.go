package auth_ldap

import "fmt"

// TODO затычка для авторизации в админке через конфиг
type ConfigAuthenticator struct {
	login       string
	password    string
	permissions []string // права, которые выдаём этому пользователю
}

func NewConfigAuthenticator(login, password string, permissions []string) *ConfigAuthenticator {
	return &ConfigAuthenticator{login: login, password: password, permissions: permissions}
}

func (a *ConfigAuthenticator) AuthenticateUser(login string, password string) (*User, []string, error) {
	if login != a.login || password != a.password {
		return nil, nil, fmt.Errorf("invalid login or password")
	}
	user := &User{
		UID:      login,
		FullName: login,
	}

	return user, a.permissions, nil
}
