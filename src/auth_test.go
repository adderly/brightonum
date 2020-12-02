package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"testing"

	"ruslanlesko/brightonum/src/dao"
	s "ruslanlesko/brightonum/src/structs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const baseURL string = "http://localhost:2525/"

var user = s.User{42, "alle", "test", "user", "test@email.com", "$2a$04$Mhlu1.a4QchlVgGQFc/0N.qAw9tsXqm1OMwjJRaPRCWn47bpsRa4S"}
var updatedUser = s.User{42, "", "", "", "updated@email.com", ""}
var user2 = s.User{-1, "sarah", "Sarah", "Lynn", "sarah@email.com", "oakheart"}
var userInfo = s.UserInfo{42, "alle", "test", "user", "test@email.com"}

func TestMain(m *testing.M) {
	setup()
	retCode := m.Run()
	os.Exit(retCode)
}

// TokenResponse encapsulates token response
type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func TestFunctional_GetByUsername(t *testing.T) {
	resp, err := http.Get(baseURL + "v1/userinfo/byusername/" + user.Username)
	assert.Nil(t, err)

	defer resp.Body.Close()

	var resultUserInfo s.UserInfo
	json.NewDecoder(resp.Body).Decode(&resultUserInfo)
	assert.Equal(t, userInfo, resultUserInfo)
}

func TestFunctional_GetById(t *testing.T) {
	resp, err := http.Get(baseURL + "v1/userinfo/byid/" + strconv.Itoa(user.ID))
	assert.Nil(t, err)

	defer resp.Body.Close()

	var resultUserInfo s.UserInfo
	json.NewDecoder(resp.Body).Decode(&resultUserInfo)
	assert.Equal(t, userInfo, resultUserInfo)
}

func TestFunctional_CreateUser(t *testing.T) {
	resp, err := http.Post(baseURL+"v1/users", "application/json", bytes.NewReader(s.U2JSON(&user)))
	assert.Nil(t, err)

	assert.Equal(t, 400, resp.StatusCode)

	resp, err = http.Post(baseURL+"v1/users", "application/json", bytes.NewReader(s.U2JSON(&user2)))
	assert.Nil(t, err)

	assert.Equal(t, 201, resp.StatusCode)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)

	assert.JSONEq(t, "{ \"id\" : 43}", string(body))
}

func TestFunctional_Token(t *testing.T) {
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodPost, baseURL+"v1/token", nil)
	assert.Nil(t, err)
	req.Header.Add("Authorization", "Basic YWxsZTpvYWtoZWFydA==")
	resp, err := client.Do(req)
	assert.Nil(t, err)

	var tokenResp TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	assert.Nil(t, err)

	assert.True(t, len(tokenResp.AccessToken) > 1)
	assert.True(t, len(tokenResp.RefreshToken) > 1)

	req, err = http.NewRequest(http.MethodPost, baseURL+"v1/token?type=refresh_token", nil)
	assert.Nil(t, err)
	req.Header.Add("Authorization", "Bearer "+tokenResp.RefreshToken)
	resp, err = client.Do(req)
	assert.Nil(t, err)

	var tokenResp2 TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResp2)
	assert.Nil(t, err)
	assert.True(t, len(tokenResp2.AccessToken) > 1)
	assert.Empty(t, tokenResp2.RefreshToken)
}

func TestFunctional_Update(t *testing.T) {
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodPost, baseURL+"v1/token", nil)
	assert.Nil(t, err)
	req.Header.Add("Authorization", "Basic YWxsZTpvYWtoZWFydA==")
	resp, err := client.Do(req)
	assert.Nil(t, err)

	var tokenResp TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	assert.Nil(t, err)

	assert.True(t, len(tokenResp.AccessToken) > 1)
	assert.True(t, len(tokenResp.RefreshToken) > 1)

	req, err = http.NewRequest(http.MethodPatch, baseURL+"v1/users/42", bytes.NewReader(s.U2JSON(&updatedUser)))
	assert.Nil(t, err)
	req.Header.Add("Authorization", "Bearer "+tokenResp.AccessToken)
	resp, err = client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func setup() {
	dao := dao.MockUserDao{}
	dao.On("GetByUsername", user.Username).Return(&user, nil)
	dao.On("GetByUsername", user2.Username).Return(nil, nil)
	dao.On("Get", user.ID).Return(&user, nil)
	dao.On("Save", mock.MatchedBy(
		func(u *s.User) bool {
			return u.Username == user2.Username && u.FirstName == user2.FirstName && u.LastName == user2.LastName
		})).Return(43)
	dao.On("Update", &updatedUser).Return(nil)

	conf := Config{PrivKeyPath: "../test_data/private.pem", PubKeyPath: "../test_data/public.pem"}
	service := AuthService{UserDao: &dao, Config: conf}

	auth := Auth{AuthService: &service}
	go auth.start()
}
