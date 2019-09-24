package model

type User struct {
	ID       int64  `json:"id" form:"ID" sql:"id"`
	NickName string `json:"nick_name" form:"NickName" sql:"nick_name"`
	UserName string `json:"user_name" form:"UserName" sql:"user_name"`
	// 密码 json不做映射
	HashPassword string `json:"-" form:"Password" sql:"password"`
}
