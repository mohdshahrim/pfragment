// user group definition, rights, policies & access control
package main

// standard terms...
// CREATE = new, add
// READ = get
// UPDATE = modify
// DELETE = remove, terminate
// OWN = self,
// USER = other user

type GroupPermission interface {
	UserPermission() bool
}

// returns brief description about usergroup
func UsergroupDefinition(usergroup string) string {
	switch usergroup {
	case "guest":
		return "The lowest user level, with many limited features and access"
	case "normal":
		return "Common user and is allowed to access most features that does not involve users management and system management"
	case "admin":
		return "Powerful user with extended privileges and able to manage other users, subsystem, and many more"
	case "master":
		return "Top-most user level and no restrictions at all. Only 1 account is allowed to exist"
	default:
		return "No usergroup is defined for this " + usergroup
	}
}

// determines eligibility to update own account password
func UpdateOwnPassword(usergroup string) bool {
	switch usergroup {
	case "guest":
		return false
	case "normal":
		return true
	case "admin":
		return true
	case "master":
		return true
	default:
		return false
	}
}

// determines eligibility to update other user's password
func UpdateUserPassword(usergroup string) bool {
	switch usergroup {
	case "guest":
		return false
	case "normal":
		return false
	case "admin":
		return true
	case "master":
		return true
	default:
		return false
	}
}

// determines eligibility to access admin pages
func AccessAdmin(usergroup string) bool {
	switch usergroup {
	case "guest":
		return false
	case "normal":
		return false
	case "admin":
		return true
	case "master":
		return true
	default:
		return false
	}
}