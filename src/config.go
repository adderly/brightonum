package main

// Config provides configuration variables
type Config struct {
	// Path to a private key
	PrivKeyPath string `long:"privkey" required:"true" description:"Path to a private key"`

	// Path to a public key
	PubKeyPath string `long:"pubkey" required:"true" description:"Path to a public key"`

	// MongoDB URL
	databaseURL string `long:"databaseDriverUrl" required:"true" description:"URL for MongoDB"`

	// Database name
	DatabaseName string `long:"databaseName" required:"true" description:"Database name"`

	// Email for password recovery (Gmail)
	Email string `long:"email" required:"true" description:"Email for password recovery (Gmail)"`

	// Password from email for password recovery
	EmailPassword string `long:"emailPassword" required:"true" description:"Password from email for password recovery"`

	// Enable debug logging
	Debug bool `long:"debug" required:"false" description:"Enable debug logging"`

	// Admin ID
	AdminID int64 `long:"adminID" required:"true" description:"Admin ID"`

	// Enable private mode
	Private bool `long:"private" required:"false" description:"Private Mode"`

	// The database driver thar will be used
	DriverName string `long:"driverName" required:"true" description:"Database driver name (mysql, mongodb, etc)"`
}
