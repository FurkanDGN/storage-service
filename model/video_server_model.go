package model

type VideoServer struct {
	ID      string `bson:"_id,omitempty"        json:"id,omitempty"`
	Address string `bson:"address,omitempty"    json:"address,omitempty"`
	Port    string `bson:"port,omitempty"       json:"port,omitempty"`
}
