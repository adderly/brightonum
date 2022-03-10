package dao

import (
	"context"
	"os"
	"os/signal"
	s "ruslanlesko/brightonum/src/structs"
	"strings"
	"syscall"

	"xorm.io/builder"
	"xorm.io/xorm"
)

// MysqlUserDao provides UserDao implementation via MongoDB
type MysqlUserDao struct {
	Db           *xorm.Engine
	DatabaseName string
	Ctx          context.Context
}

// NewMysqlUserDao creates instance of MysqlUserDao
func NewMysqlUserDao(driverUrl string, databaseName string) *MysqlUserDao {
	ctx, cancel := context.WithCancel(context.Background())
	dbClient, err := xorm.NewEngine(driverUrl, databaseName)

	if err != nil {
		logger.Logf("ERROR Failed to dial mongo url: '%s'", driverUrl)
		panic(err)
	}

	if err = dbClient.Sync2(new(s.User)); err != nil {
		logger.Logf("orm failed to initialized User table: %v", err)
	}
	logger.Logf("INFO Connected to SQLDb")

	sigChan := make(chan os.Signal)
	go func() {
		for range sigChan {
			logger.Logf("INFO disconnecting from SQLDb")
			dbClient.Close()
			logger.Logf("INFO disconnected from SQLDb")
			cancel()
		}
	}()
	signal.Notify(sigChan, syscall.SIGTERM)

	return &MysqlUserDao{Db: dbClient, DatabaseName: databaseName, Ctx: ctx}
}

// Save saves user in SQLDb.
// Implemented to retry insertion several times if another thread inserts document between
// calculation of new id and insertion into collection.
func (d *MysqlUserDao) Save(u *s.User) int64 {
	u.Username = strings.ToLower(u.Username)

	var user interface{}
	user = u

	affected, err := d.Db.Insert(&user)

	if err != nil {
		return -1 //TODO: something happened
	}

	return affected
}

// GetByUsername extracts user by username
func (d *MysqlUserDao) GetByUsername(username string) (*s.User, error) {
	result := &s.User{}

	q := builder.Expr("username = ?", strings.ToLower(username))
	err := d.Db.Where(q).Find(result)

	if err != nil {
		logger.Logf("ERROR %s", err)
		return nil, err
	}

	return result, nil
}

// GetByEmail returns nil when user is not found
// Returns error if data access error occured
func (d *MysqlUserDao) GetByEmail(email string) (*s.User, error) {
	result := &s.User{}

	q := builder.Expr("email = ?", strings.ToLower(email))
	err := d.Db.Where(q).Find(result)

	if err != nil {
		logger.Logf("ERROR %s", err)
		return nil, err
	}

	return result, nil
}

// Get returns user by id
func (d *MysqlUserDao) Get(id int64) (*s.User, error) {
	result := &s.User{}

	q := builder.Expr("ID = ?", id)
	err := d.Db.Where(q).Find(result)

	if err != nil {
		logger.Logf("ERROR %s", err)
		return nil, err
	}

	return result, nil
}

// GetAll extracts all users
func (d *MysqlUserDao) GetAll() (*[]s.User, error) {
	result := []s.User{}

	err := d.Db.Find(result)

	if err != nil {
		logger.Logf("ERROR %s", err)
		return nil, err
	}

	return &result, nil
}

// Update updates user if exists
func (d *MysqlUserDao) Update(u *s.User) error {

	updatedUser := &s.User{}

	if u.FirstName != "" {
		updatedUser.FirstName = u.FirstName
	}
	if u.LastName != "" {
		updatedUser.LastName = u.LastName
	}
	if u.Email != "" {
		updatedUser.Email = u.Email
	}
	if u.Password != "" {
		updatedUser.Password = u.Password
	}

	_, err := d.Db.ID(u.ID).Update(updatedUser)

	return err
}

// SetRecoveryCode sets password recovery code for user id
func (d *MysqlUserDao) SetRecoveryCode(id int64, code string) error {

	user := &s.User{}
	user.RecoveryCode = code
	user.ResettingCode = " "
	_, err := d.Db.ID(id).Update(user)

	return err
}

// GetRecoveryCode extracts recovery code for user id
func (d *MysqlUserDao) GetRecoveryCode(id int64) (string, error) {
	var user s.User
	q := builder.Expr("ID = ", id)
	err := d.Db.Where(q).Find(&user)
	return user.RecoveryCode, err
}

// SetResettingCode sets resetting code and removes recovery one
func (d *MysqlUserDao) SetResettingCode(id int64, code string) error {
	user := &s.User{}
	user.ResettingCode = code
	user.RecoveryCode = " "
	_, err := d.Db.ID(id).Update(user)
	return err
}

// GetResettingCode extracts resetting code for user id
func (d *MysqlUserDao) GetResettingCode(id int64) (string, error) {
	var user s.User
	q := builder.Expr("ID = ", id)
	err := d.Db.Where(q).Find(&user)
	return user.ResettingCode, err
}

// ResetPassword updates password and removes resetting code
func (d *MysqlUserDao) ResetPassword(id int64, passwordHash string) error {
	user := &s.User{}
	user.Password = passwordHash
	user.ResettingCode = " "
	_, err := d.Db.ID(id).Update(user)
	return err
}

// DeleteById deletes user by id
func (d *MysqlUserDao) DeleteById(id int64) error {
	q := builder.Expr("ID = ?", id)
	_, err := d.Db.Where(q).Delete()
	return err
}
