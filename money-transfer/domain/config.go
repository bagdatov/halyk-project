package domain

import "time"

// Config ...
type Config struct {
	BindAddr string `toml:"bind_addr"`

	HostDB     string `toml:"db_host"`
	PortDB     string `toml:"db_port"`
	NameDB     string `toml:"db_name"`
	UserDB     string `toml:"db_user"`
	PasswordDB string `toml:"db_password"`

	AccessTokenSecret  string   `toml:"access_token_secret"`
	RefreshTokenSecret string   `toml:"refresh_token_secret"`
	AccessTokenTTL     duration `toml:"access_token_ttl"`
	RefreshTokenTTL    duration `toml:"refresh_token_ttl"`
}

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

func NewConfig() *Config {
	return &Config{
		BindAddr: ":7575",

		HostDB:     "localhost",
		PortDB:     ":5432",
		NameDB:     "postgres",
		UserDB:     "postgres",
		PasswordDB: "postgres",

		AccessTokenTTL:  duration{10 * time.Minute},
		RefreshTokenTTL: duration{1 * time.Hour},
	}
}
