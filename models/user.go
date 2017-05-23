package models

type User struct {
	AccountID   int64
	GUID        string
	GameID      string
	Channel     string
	ServerID    string
	Salt        string
	GameVersion string
}

type SMS struct {
	Type        string
	Code        string
	ExpiredTime int64
}
