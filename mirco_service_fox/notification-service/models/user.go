package models

type User struct {
	ID        uint
	Name      string
	Telephone string
	Password  string
	Role      string
	AvatarUrl string
	Email     string
}

func GetUserFromID(ID uint) User {
	// TODO:wait for done

	user := User{
		ID: ID,
	}
	return user
}
