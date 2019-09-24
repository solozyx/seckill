package dao

import (
	"database/sql"
	"errors"
	"strconv"

	"github.com/solozyx/seckill/datasource"
	"github.com/solozyx/seckill/model"
)

type IUser interface {
	Conn() error
	// 用户系统 user_name 唯一
	Select(userName string) (user *model.User, err error)
	Insert(user *model.User) (userId int64, err error)
}

type UserManager struct {
	mysqlConn *sql.DB
}

func NewUserManager(db *sql.DB) IUser {
	return &UserManager{db}
}

func (u *UserManager) Conn() (err error) {
	if u.mysqlConn == nil {
		mysql, err := datasource.NewMysqlConn()
		if err != nil {
			return err
		}
		u.mysqlConn = mysql
	}
	return
}

func (u *UserManager) Select(userName string) (user *model.User, err error) {
	if userName == "" {
		return &model.User{}, errors.New("用户名不能为空！")
	}
	if err = u.Conn(); err != nil {
		return &model.User{}, err
	}

	// 占位符 ? 避免恶意用户输入非法字符攻击数据库
	sql := `select * from user where user_name=?`
	rows, err := u.mysqlConn.Query(sql, userName)
	if err != nil {
		return &model.User{}, err
	}
	defer rows.Close()

	result := datasource.GetResultRow(rows)
	if len(result) == 0 {
		return &model.User{}, errors.New("用户不存在！")
	}

	user = &model.User{}
	// 根据 User 结构体的 `sql:"user_name"` 标签 实现 map --> User struct
	datasource.DataToStructByTagSql(result, user)
	return
}

func (u *UserManager) Insert(user *model.User) (userId int64, err error) {
	if err = u.Conn(); err != nil {
		return
	}

	sql := `insert user set nick_name=?,user_name=?,password=?`
	stmt, err := u.mysqlConn.Prepare(sql)
	if err != nil {
		return
	}
	defer stmt.Close()

	result, err := stmt.Exec(user.NickName, user.UserName, user.HashPassword)
	if err != nil {
		return
	}
	return result.LastInsertId()
}

func (u *UserManager) SelectById(userId int64) (user *model.User, err error) {
	if err = u.Conn(); err != nil {
		return &model.User{}, err
	}

	sql := "select * from user where id=" + strconv.FormatInt(userId, 10)
	row, err := u.mysqlConn.Query(sql)
	if err != nil {
		return &model.User{}, err
	}
	defer row.Close()

	result := datasource.GetResultRow(row)
	if len(result) == 0 {
		return &model.User{}, errors.New("用户不存在！")
	}
	user = &model.User{}
	datasource.DataToStructByTagSql(result, user)
	return
}
