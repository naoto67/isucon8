package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

var (
	USER_ID_KEY         = "uid:"
	USER_LOGIN_NAME_KEY = "uln:"

	ErrLoginNameEx = errors.New("user found")
)

func InitUserCache() error {
	var users []User
	err := db.Select(&users, "SELECT * FROM users")
	if err != nil {
		return err
	}
	dict := map[string][]byte{}
	for _, v := range users {
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		key := fmt.Sprintf("%s%d", USER_ID_KEY, v.ID)
		dict[key] = data
		key = fmt.Sprintf("%s%s", USER_LOGIN_NAME_KEY, v.LoginName)
		dict[key] = data
	}
	err = cacheClient.MultiSet(dict)
	return err
}

func RegisterUser(nickname, loginName, password string) (*User, error) {
	key := fmt.Sprintf("%s%s", USER_LOGIN_NAME_KEY, loginName)
	d, _ := cacheClient.SingleGet(key)
	if d != nil {
		return nil, ErrLoginNameEx
	}
	sum := sha256.Sum256([]byte(password))
	passHash := hex.EncodeToString(sum[:])
	fmt.Println("RegisterUser: passHash", passHash)

	tx := db.MustBegin()
	res, err := tx.Exec("INSERT INTO users (login_name, pass_hash, nickname) VALUES (?, ?, ?)", loginName, passHash, nickname)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	userID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	user := &User{
		ID:        userID,
		Nickname:  nickname,
		LoginName: loginName,
		PassHash:  passHash,
	}
	data, err := json.Marshal(user)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	key = fmt.Sprintf("%s%d", USER_ID_KEY, userID)
	err = cacheClient.SingleSet(key, data)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	key = fmt.Sprintf("%s%s", USER_LOGIN_NAME_KEY, loginName)
	err = cacheClient.SingleSet(key, data)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return user, nil
}
