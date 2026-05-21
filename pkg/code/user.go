package code

// User 10100–10199

const (
	UserDevAuthDisabled Code = 10101
	UserIDRequired      Code = 10102
	UserNotFound        Code = 10103
	UserLoginNotReady   Code = 10104
	UserDeviceRequired  Code = 10105
	UserRegisterFailed  Code = 10106
)

func init() {
	register(UserDevAuthDisabled, "dev_auth_disabled", "dev auth disabled")
	register(UserIDRequired, "user_id_required", "user_id required")
	register(UserNotFound, "user_not_found", "user not found")
	register(UserLoginNotReady, "login_not_ready", "login not implemented yet")
	register(UserDeviceRequired, "device_id_required", "device_id required")
	register(UserRegisterFailed, "register_failed", "register failed")
}
