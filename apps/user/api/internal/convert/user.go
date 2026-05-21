package convert

import (
	"im/pkg/models"
	"im/apps/user/api/internal/types"
)

func UserToAPI(u *models.User) types.User {
	if u == nil {
		return types.User{}
	}
	return types.User{
		Id:        u.ID,
		Username:  u.Username,
		Nickname:  u.Nickname,
		AvatarUrl: u.AvatarURL,
	}
}
