package model

type UserToken struct {
	Token string `json:"token"`
}

type UserInfo struct {
	Roles        []string `json:"token"`
	Introduction string   `json:"introduction"`
	Avatar       string   `json:"avatar"`
	Name         string   `json:"name"`
}
