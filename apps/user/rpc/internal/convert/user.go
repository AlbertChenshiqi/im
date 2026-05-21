package convert

import (
	"im/pkg/models"
	"im/apps/user/rpc/user"
)

func UserToRPC(u *models.User) *user.UserInfo {
	if u == nil {
		return nil
	}
	return &user.UserInfo{
		Id:        u.ID,
		Username:  u.Username,
		Nickname:  u.Nickname,
		AvatarUrl: u.AvatarURL,
	}
}
