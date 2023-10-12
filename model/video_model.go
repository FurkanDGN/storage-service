package model

type Video struct {
	ID      string `bson:"_id,omitempty" json:"id,omitempty"`
	Title   string `bson:"title,omitempty" json:"title,omitempty"`
	VideoUrl string `bson:"video_url,omitempty" json:"videoUrl,omitempty"`
}