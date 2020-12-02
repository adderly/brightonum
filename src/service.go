package main

import (
	"fmt"
	"io/ioutil"
	"ruslanlesko/brightonum/src/crypto"
	"ruslanlesko/brightonum/src/dao"
	st "ruslanlesko/brightonum/src/structs"

	"time"

	"github.com/dgrijalva/jwt-go"
)

// AuthService provides all auth operations
type AuthService struct {
	UserDao dao.UserDao
	Config  Config
}

// CreateUser creates new User
func (s *AuthService) CreateUser(u *st.User) error {
	logger.Logf("DEBUG creating user")

	uname := u.Username

	alreadyExists, err := s.usernameExists(uname)
	if err != nil {
		return st.AuthError{Msg: err.Error(), Status: 500}
	}
	if alreadyExists {
		logger.Logf("WARN Username %s already exists", uname)
		return st.AuthError{Msg: "Username already exists", Status: 400}
	}

	hashedPassword, err := crypto.Hash(u.Password)
	if err != nil {
		logger.Logf("ERROR Failed to hash password, %s", err.Error())
		return err
	}

	u.Password = hashedPassword
	ID := s.UserDao.Save(u)
	if ID < 0 {
		return st.AuthError{Msg: "Cannot save user", Status: 500}
	}
	u.ID = ID
	return nil
}

// UpdateUser updates existing user
func (s *AuthService) UpdateUser(u *st.User, token string) error {
	logger.Logf("DEBUG Updating user with id %d", u.ID)

	tokenUser, valid := s.validateToken(token)
	if !valid || tokenUser.ID != u.ID {
		return st.AuthError{Msg: "Invalid token", Status: 401}
	}

	if !validateUpdatePayload(u) {
		return st.AuthError{Msg: "Invalid Update payload", Status: 400}
	}

	userExists, err := s.userExists(u.ID)
	if err != nil {
		return st.AuthError{Msg: err.Error(), Status: 500}
	}
	if !userExists {
		return st.AuthError{Msg: "User does not exist", Status: 404}
	}

	err = s.UserDao.Update(u)
	if err != nil {
		return st.AuthError{Msg: err.Error(), Status: 500}
	}

	return nil
}

func (s *AuthService) usernameExists(username string) (bool, error) {
	u, err := s.UserDao.GetByUsername(username)
	return u != nil, err
}

func validateUpdatePayload(u *st.User) bool {
	return u.ID > 0 && u.Username == "" && u.Password == ""
}

func (s *AuthService) userExists(id int) (bool, error) {
	u, err := s.UserDao.Get(id)
	if err != nil {
		return false, err
	}
	return u != nil, nil
}

// BasicAuthToken issues new token by username and password
func (s *AuthService) BasicAuthToken(username, password string) (string, string, error) {
	user, err := s.UserDao.GetByUsername(username)

	if err != nil {
		return "", "", st.AuthError{Msg: "Cannot extract user", Status: 500}
	}

	if user == nil || !crypto.Match(password, user.Password) {
		return "", "", st.AuthError{Msg: "Username or password is wrong", Status: 403}
	}

	tokenString, err := s.issueAccessToken(user)
	if err != nil {
		return "", "", err
	}

	refreshTokenString, err := s.issueRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	return tokenString, refreshTokenString, nil
}

func (s *AuthService) issueAccessToken(user *st.User) (string, error) {
	if user == nil {
		return "", st.AuthError{Msg: "User is missing", Status: 403}
	}

	keyData, err := ioutil.ReadFile(s.Config.PrivKeyPath)
	if err != nil {
		return "", st.AuthError{Msg: err.Error(), Status: 500}
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		return "", st.AuthError{Msg: err.Error(), Status: 500}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub":    user.Username,
		"userId": user.ID,
		"exp":    time.Now().Add(time.Hour).UTC().Unix(),
	})

	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", st.AuthError{Msg: err.Error(), Status: 500}
	}

	return tokenString, nil
}

func (s *AuthService) issueRefreshToken(user *st.User) (string, error) {
	keyData, err := ioutil.ReadFile(s.Config.PrivKeyPath)
	if err != nil {
		return "", st.AuthError{Msg: err.Error(), Status: 500}
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		return "", st.AuthError{Msg: err.Error(), Status: 500}
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": user.Username,
		"exp": time.Now().AddDate(1, 0, 0).UTC().Unix(),
	})

	refreshTokenString, err := refreshToken.SignedString(key)
	if err != nil {
		return "", st.AuthError{Msg: err.Error(), Status: 500}
	}

	return refreshTokenString, nil
}

// RefreshToken refreshes existing token
func (s *AuthService) RefreshToken(t string) (string, error) {
	u, ok := s.validateToken(t)
	if ok {
		accessToken, err := s.issueAccessToken(u)
		return accessToken, err
	}
	return "", st.AuthError{Msg: "Refresh token is not valid", Status: 403}
}

func (s *AuthService) validateToken(t string) (*st.User, bool) {
	keyData, err := ioutil.ReadFile(s.Config.PubKeyPath)
	if err != nil {
		return nil, false
	}
	key, err := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if err != nil {
		logger.Logf("WARN %s", err.Error())
		return nil, false
	}

	token, err := jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	})
	if err != nil {
		logger.Logf("WARN %s", err.Error())
		return nil, false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		u, err := s.UserDao.GetByUsername(fmt.Sprintf("%s", claims["sub"]))
		if err != nil {
			return nil, false
		}
		return u, true
	}
	return nil, false
}

// GetUserByToken returns user by token
func (s *AuthService) GetUserByToken(t string) (*st.User, error) {
	keyData, err := ioutil.ReadFile(s.Config.PubKeyPath)
	if err != nil {
		return nil, st.AuthError{Msg: err.Error(), Status: 500}
	}
	key, err := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if err != nil {
		return nil, st.AuthError{Msg: err.Error(), Status: 500}
	}

	token, err := jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	})
	if err != nil {
		return nil, st.AuthError{Msg: err.Error(), Status: 400}
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		u, err := s.UserDao.GetByUsername(fmt.Sprintf("%s", claims["sub"]))
		if err != nil {
			return nil, st.AuthError{Msg: err.Error(), Status: 500}
		}
		return u, nil
	}
	return nil, nil
}

// GetUserById returns user info for specific id
func (s *AuthService) GetUserById(id int) (*st.UserInfo, error) {
	u, err := s.UserDao.Get(id)
	if err != nil {
		return nil, st.AuthError{Msg: err.Error(), Status: 500}
	}
	if u == nil {
		return nil, nil
	}
	return &st.UserInfo{ID: u.ID, Username: u.Username, FirstName: u.FirstName, LastName: u.LastName, Email: u.Email}, nil
}

// GetUserById returns user info for username
func (s *AuthService) GetUserByUsername(username string) (*st.UserInfo, error) {
	u, err := s.UserDao.GetByUsername(username)
	if err != nil {
		return nil, st.AuthError{Msg: err.Error(), Status: 500}
	}
	if u == nil {
		return nil, nil
	}
	return mapToUserInfo(u), nil
}

// GetUsers returns all users info
func (s *AuthService) GetUsers() (*[]st.UserInfo, error) {
	us, err := s.UserDao.GetAll()
	if err != nil {
		return nil, st.AuthError{Msg: err.Error(), Status: 500}
	}
	return mapToUserInfoList(us), nil
}

func mapToUserInfoList(us *[]st.User) *[]st.UserInfo {
	result := []st.UserInfo{}
	for _, u := range *us {
		result = append(result, *mapToUserInfo(&u))
	}
	return &result
}

func mapToUserInfo(u *st.User) *st.UserInfo {
	return &st.UserInfo{ID: u.ID, Username: u.Username, FirstName: u.FirstName, LastName: u.LastName, Email: u.Email}
}
