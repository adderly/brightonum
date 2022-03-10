package dao

import (
	"ruslanlesko/brightonum/src/structs"

	"github.com/stretchr/testify/mock"
)

// MockUserDao for testing only
type MockUserDao struct {
	mock.Mock
}

func (m *MockUserDao) Save(u *structs.User) int64 {
	return int64(m.Called(u).Int(0))
}

func (m *MockUserDao) GetByUsername(uname string) (*structs.User, error) {
	provided := m.Called(uname).Get(0)
	err := m.Called(uname).Get(1)
	var castedErr error = nil
	if err != nil {
		castedErr = err.(error)
	}
	if provided == nil {
		return nil, castedErr
	}
	return provided.(*structs.User), castedErr
}

func (m *MockUserDao) GetByEmail(uname string) (*structs.User, error) {
	provided := m.Called(uname).Get(0)
	err := m.Called(uname).Get(1)
	var castedErr error = nil
	if err != nil {
		castedErr = err.(error)
	}
	if provided == nil {
		return nil, castedErr
	}
	return provided.(*structs.User), castedErr
}

func (m *MockUserDao) Get(id int64) (*structs.User, error) {
	provided := m.Called(id).Get(0)
	err := m.Called(id).Get(1)
	var castedErr error = nil
	if err != nil {
		castedErr = err.(error)
	}
	if provided == nil {
		return nil, castedErr
	}
	return provided.(*structs.User), castedErr
}

func (m *MockUserDao) GetAll() (*[]structs.User, error) {
	provided := m.Called().Get(0)
	err := m.Called().Get(1)
	var castedErr error = nil
	if err != nil {
		castedErr = err.(error)
	}
	if provided == nil {
		return nil, castedErr
	}
	return provided.(*[]structs.User), castedErr
}

func (m *MockUserDao) Update(u *structs.User) error {
	err := m.Called(u).Get(0)
	var castedErr error = nil
	if err != nil {
		castedErr = err.(error)
	}
	return castedErr
}

func (m *MockUserDao) SetRecoveryCode(id int64, code string) error {
	err := m.Called(id, code).Get(0)
	var castedErr error = nil
	if err != nil {
		castedErr = err.(error)
	}
	return castedErr
}

func (m *MockUserDao) GetRecoveryCode(id int64) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

func (m *MockUserDao) SetResettingCode(id int64, code string) error {
	return m.Called(id, code).Error(0)
}

func (m *MockUserDao) GetResettingCode(id int64) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

func (m *MockUserDao) ResetPassword(id int64, passwordHash string) error {
	return m.Called(id, passwordHash).Error(0)
}

func (m *MockUserDao) DeleteById(id int64) error {
	return m.Called(id).Error(0)
}
