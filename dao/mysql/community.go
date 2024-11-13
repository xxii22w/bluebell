package mysql

import (
	"bluebell/models"
	"database/sql"

	"go.uber.org/zap"
)

func GetCommunityList() (communityList []*models.Community, err error) {
	sqlStr := "select community_id,community_name from community"
	if err := db.Select(&communityList, sqlStr); err != nil {
		if err == sql.ErrNoRows {
			zap.L().Warn("there is no community in db")
			err = nil
		}
	}
	return
}

// GetCommunityDetailByID 根据Id查询社区详情
func GetCommunityDetailByID(id int64) (commity *models.CommunityDetail, err error) {
	commity = new(models.CommunityDetail)
	sqlStr := "select community_id,community_name,introduction,create_time from community where community_id = ?"
	if err := db.Get(commity, sqlStr, id); err != nil {
		if err == sql.ErrNoRows {
			err = ErrorInvalidID
		}
	}
	return commity, err
}
