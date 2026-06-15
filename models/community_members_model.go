package models

import "time"

const (
	MemberRoleMember     = "member"
	MemberRoleModerator  = "moderator"
	MemberRoleAdmin      = "admin"

	MemberStatusActive   = "active"
	MemberStatusBanned   = "banned"
	MemberStatusPending  = "pending"
)

type CommunityMember struct {
	ID         uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CommunityID uint     `gorm:"column:community_id;not null;index;uniqueIndex:idx_community_members_community_user" json:"community_id"`
	Community  Community `gorm:"foreignKey:CommunityID;references:ID" json:"community,omitempty"`
	UserID     uint      `gorm:"column:user_id;not null;index;uniqueIndex:idx_community_members_community_user" json:"user_id"`
	User       User      `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Role       string    `gorm:"column:role;type:varchar(50);not null;default:member" json:"role"`
	Status     string    `gorm:"column:status;type:varchar(50);not null;default:active;index" json:"status"`
	JoinedAt   time.Time `gorm:"column:joined_at;autoCreateTime" json:"joined_at"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (CommunityMember) TableName() string {
	return "community_members"
}
