package config

import (
	"bytes"
)

type Config struct {
	User struct {
		id       int 		`yaml:"id"`
		email    string		`yaml:"email"`
		password string		`yaml:"password"`
	}
	Database struct {
		host 	string		`yaml:"host"`
		port 	int			`yaml:"port"`
		name 	string		`yaml:"name"`
	}
	Profile struct {
		profileId int		`yaml:"profileId"`
		firstName string	`yaml:"firstName"`
		lastName  string	`yaml:"lastName"`
		height    int		`yaml:"height"`
		birthday  struct {
			year  int		`yaml:"year"`
			month int		`yaml:"month"`
			day   int		`yaml:"day"`
		}
		avatar      bytes.Buffer	`yaml:"avatar"`
		description string			`yaml:"description"`
		location    string			`yaml:"location"`
		interests   []string		`yaml:"interests"`
		likedBy     []int			`yaml:"likedBy"`
		preferences struct {
			preferencesId int		`yaml:"preferencesId"`
			interests     []string	`yaml:"interests"`
			location      string	`yaml:"location"`
			age           struct {
				from int	`yaml:"from"`
				to   int	`yaml:"to"`
			}
		}
	}
}
