package model

type FtpServer struct {
	ID       string `bson:"_id,omitempty"        json:"id,omitempty"`
	Address  string `bson:"address,omitempty"    json:"address,omitempty"`
	Port     string `bson:"port,omitempty"       json:"port,omitempty"`
	Username string `bson:"username,omitempty"   json:"username,omitempty"`
	Password string `bson:"password,omitempty"   json:"password,omitempty"`
}