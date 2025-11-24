package userData

import (
	commonUserData "github.com/luskaner/ageLANServer/launcher-common/userData"
)

func Metadata(gameId string) Data {
	data := Data{}
	err, metadatas := commonUserData.Metadatas(gameId)
	if err != nil {
		return data
	}
	for metadata := range metadatas.Iter() {
		if metadata.Type == commonUserData.TypeActive {
			data.Path = metadata.Path
			break
		}
	}
	return data
}
