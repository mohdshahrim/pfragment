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

func UsergroupPermission(permission string, usergroup string) bool {
	switch permission {
	case "update_own_password":
		return UpdateOwnPassword(usergroup)
	case "update_user_password":
		return UpdateUserPassword(usergroup)
	case "access_admin":
		return AccessAdmin(usergroup)
	case "access_itdb":
		return AccessITDB(usergroup)
	default:
		return false
	}
}

// returns brief description about usergroup
func UsergroupDefinition(usergroup string) string {
	switch usergroup {
	case "normal":
		return "Common user and is allowed to access most features that does not involve users management and system management"
	case "admin":
		return "Powerful user with extended privileges and able to manage other users, subsystem, and many more"
	default:
		return "No usergroup is defined for this " + usergroup
	}
}

// determines eligibility to update own account password
func UpdateOwnPassword(usergroup string) bool {
	switch usergroup {
	case "normal":
		return true
	case "admin":
		return true
	default:
		return false
	}
}

// determines eligibility to update other user's password
func UpdateUserPassword(usergroup string) bool {
	switch usergroup {
	case "normal":
		return false
	case "admin":
		return true
	default:
		return false
	}
}

// determines eligibility to access admin pages
func AccessAdmin(usergroup string) bool {
	switch usergroup {
	case "normal":
		return false
	case "admin":
		return true
	default:
		return false
	}
}

// determines the eligibility to access itdb system
func AccessITDB(usergroup string) bool {
	switch usergroup {
	case "normal":
		return false
	case "admin":
		return true
	default:
		return false
	}
}