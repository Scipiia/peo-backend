package auth_ldap

import (
	"fmt"
	"log"
	"strings"
	"vue-golang/internal/config"

	"github.com/go-ldap/ldap/v3"
)

type LDAPClient struct {
	config config.LDAPConfig
}

func NewLDAPClient(cfg config.LDAPConfig) *LDAPClient {
	return &LDAPClient{config: cfg}
}

func (l *LDAPClient) AuthenticateUser(username, password string) (*User, []string, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)

	conn, err := ldap.DialURL(l.config.URL)
	if err != nil {
		return nil, nil, fmt.Errorf("LDAP connection error: %w", err)
	}
	defer conn.Close()

	err = conn.Bind(l.config.AdminDN, l.config.AdminPassword)
	if err != nil {
		return nil, nil, fmt.Errorf("LDAP connect error (URL: %s): %w", l.config.URL, err)
	}

	//searchRequest := ldap.NewSearchRequest(
	//	l.config.UserSearchBase,
	//	ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
	//	fmt.Sprintf(l.config.UserFilter, ldap.EscapeFilter(username)),
	//	[]string{"dn", "cn", "uid", "mail"},
	//	nil,
	//)

	//sr, err := conn.Search(searchRequest)
	//if err != nil || len(sr.Entries) != 1 {
	//	return nil, nil, fmt.Errorf("user not found")
	//}

	filter := fmt.Sprintf(l.config.UserFilter, ldap.EscapeFilter(username))
	searchRequest := ldap.NewSearchRequest(
		l.config.UserSearchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		[]string{"dn", "cn", "uid", "mail"},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return nil, nil, fmt.Errorf("LDAP search error (filter: %s): %w", filter, err)
	}
	if len(sr.Entries) == 0 {
		return nil, nil, fmt.Errorf("user not found in LDAP (filter: %s)", filter)
	}
	if len(sr.Entries) > 1 {
		return nil, nil, fmt.Errorf("multiple users found for filter: %s", filter)
	}

	userEntry := sr.Entries[0]
	user := &User{
		DN:       userEntry.DN,
		UID:      userEntry.GetAttributeValue("uid"),
		FullName: userEntry.GetAttributeValue("cn"),
		Email:    userEntry.GetAttributeValue("mail"),
	}

	//user := &User{
	//	UID:        sr.Entries[0].DN,
	//	FullName:   sr.Entries[0].GetAttributeValue("uid"),
	//	Permission: sr.Entries[0].GetAttributeValue("cn"),
	//	DN:         sr.Entries[0].GetAttributeValue("mail"),
	//}

	authConn, err := ldap.DialURL(l.config.URL)
	if err != nil {
		return nil, nil, fmt.Errorf("LDAP connect error for auth: %w", err)
	}
	defer authConn.Close()

	err = authConn.Bind(user.DN, password)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid password for DN [%s]: %w", user.DN, err)
	}

	groupSearch := ldap.NewSearchRequest(
		l.config.GroupSearchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(l.config.GroupFilter, ldap.EscapeFilter(username)),
		[]string{"cn"},
		nil,
	)

	groupSr, err := conn.Search(groupSearch)
	if err != nil {
		log.Printf("Warning: group search error: %v", err)
		return user, []string{}, nil
	}

	groups := []string{}
	for _, entry := range groupSr.Entries {
		groups = append(groups, entry.GetAttributeValue("cn"))
	}

	return user, groups, nil
}
