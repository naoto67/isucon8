package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"unsafe"
)

var (
	USER_ID_KEY         = "uid:"
	USER_LOGIN_NAME_KEY = "uln:"

	ErrLoginNameEx = errors.New("user found")
)

func RegisterUser(nickname, loginName, password string) (*User, error) {
	key := fmt.Sprintf("%s:%s", USER_LOGIN_NAME_KEY, loginName)
	_, err := cacheClient.SingleGet(key)
	if err != nil {
		return nil, ErrLoginNameEx
	}
	sum := sha256.Sum256([]byte(password))

	tx, err := db.Begin()
	res, err := tx.Exec("INSERT INTO users (login_name, pass_hash, nickname) VALUES (?, ?, ?)", loginName, sum, nickname)
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
		PassHash:  *(*string)(unsafe.Pointer(&sum)),
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
