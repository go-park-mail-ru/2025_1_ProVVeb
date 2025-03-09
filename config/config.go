package config

type Config struct {
	User     User     `yaml:"user" json:"user"`
	Database Database `yaml:"database" json:"database"`
	Profile  Profile  `yaml:"profile" json:"profile"`
}

type User struct {
	Id       int    `yaml:"id" json:"id"`
	Login    string `yaml:"login" json:"login"`
	Password string `yaml:"password" json:"password"`
}

type Database struct {
	host string `yaml:"host"`
	port int    `yaml:"port"`
	name string `yaml:"name"`
}

type Profile struct {
	ProfileId int    `yaml:"profileId" json:"profileId"`
	FirstName string `yaml:"firstName" json:"firstName"`
	LastName  string `yaml:"lastName" json:"lastName"`
	Height    int    `yaml:"height" json:"height"`
	Birthday  struct {
		Year  int `yaml:"year" json:"year"`
		Month int `yaml:"month" json:"month"`
		Day   int `yaml:"day" json:"day"`
	}
	Avatar      string   `yaml:"avatar" json:"avatar"`
	Description string   `yaml:"description" json:"description"`
	Location    string   `yaml:"location" json:"location"`
	Interests   []string `yaml:"interests" json:"interests"`
	LikedBy     []int    `yaml:"likedBy" json:"likedBy"`
	Preferences struct {
		PreferencesId int      `yaml:"preferencesId" json:"preferencesId"`
		Interests     []string `yaml:"interests" json:"interests"`
		Location      string   `yaml:"location" json:"location"`
		Age           struct {
			From int `yaml:"from" json:"from"`
			To   int `yaml:"to" json:"to"`
		}
	}
}
