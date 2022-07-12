package model

import "strconv"

type UserID int

func (u UserID) String() string {
	return strconv.Itoa(int(u))
}

type User struct {
	Username string
	UserID   UserID
}
