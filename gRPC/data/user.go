package data

import (
	userpb "github.com/sleeg00/gRPC/protos/v1/user"
)

var UserData = []*userpb.UserMessage{
	{
		UserId:      "1",
		Name:        "sleeg",
		PhoneNumber: "01085988556",
		Age:         22,
	},
	{
		UserId:      "2",
		Name:        "Maix",
		PhoneNumber: "01082318556",
		Age:         25,
	},
}
