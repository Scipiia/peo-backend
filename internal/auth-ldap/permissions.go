package auth_ldap

func MapGroupsToPermission(groups []string) []string {
	permissions := []string{}

	for _, group := range groups {
		switch group {
		case "normirovshiki":
			permissions = append(permissions, "norms:read", "norms:create", "norms:edit")
		case "tehnologi":
			permissions = append(permissions, "norms:read", "norms:approve", "catalogs:edit")
		case "admin":
			permissions = append(permissions, "system:users:manage", "system:roles:manage")
		}
	}
	return permissions
}
