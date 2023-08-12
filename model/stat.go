package model

type UsersStat struct {
	UsersCount        uint64 `mapstructure:"users_count" json:"users_count"`
	SubscriptionCount uint64 `mapstructure:"sub_count" json:"sub_count"`
}
