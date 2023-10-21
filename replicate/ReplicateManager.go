package replicate

import (
	"slices"
	"videohub/model"
	"videohub/util"
)

/*
-> Get the server id (from config or something like that)
-> Get all video servers from MongoDB
-> Get all local video ids
-> Get 'replicates' from MongoDB of all video ids
-> Find missing replicates of each video and
queue them as ReplicateTask
*/

func QueueMissingReplicates(mongoDb *util.MongoDB) error {
	videos, err := mongoDb.GetAllVideos()
	if err != nil {
		return err
	}
	var ownVideos []model.Video

	for _, video := range videos {
		if slices.Contains(video.Replicates, "anan") {
			ownVideos = append(ownVideos, video)
		}
	}

	return nil
}
