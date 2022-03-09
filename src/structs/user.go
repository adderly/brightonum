package structs

import (
	"encoding/json"
)

// User structure
type User struct {
	ID            int64  `bson:"_id" xorm:"varchar(200)"`
	Username      string `bson:"username" xorm:"varchar(50)"`
	FirstName     string `bson:"firstName" xorm:"varchar(50)"`
	LastName      string `bson:"lastName" xorm:"varchar(50)"`
	Email         string `bson:"email" xorm:"varchar(50)"`
	Password      string `bson:"password" xorm:"varchar(60)"`
	InviteCode    string `bson:"inviteCode" xorm:"varchar(60)"`
	RecoveryCode  string `bson:"recoveryCode" xorm:"varchar(60)"`
	ResettingCode string `bson:"resettingCode" xorm:"varchar(60)"`
}

// UserInfo structure
type UserInfo struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

func U2JSON(u *User) []byte {
	data, _ := json.Marshal(u)
	return data
}

func UI2JSON(u *UserInfo) []byte {
	data, _ := json.Marshal(u)
	return data
}

func UL2JSON(us *[]UserInfo) []byte {
	data, _ := json.Marshal(us)
	return data
}
