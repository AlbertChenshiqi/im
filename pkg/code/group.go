package code

// Group 10300–10399

const (
	GroupMembersNotFound Code = 10301
	GroupTooManyMembers  Code = 10302
)

func init() {
	register(GroupMembersNotFound, "members_not_found", "one or more member users do not exist")
	register(GroupTooManyMembers, "too_many_members", "group member limit exceeded")
}
