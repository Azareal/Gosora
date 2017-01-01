package main

type Forum struct
{
	ID int
	Name string
	Active bool
	LastTopic string
	LastTopicID int
	LastReplyer string
	LastReplyerID int
	LastTopicTime string
}

type ForumSimple struct
{
	ID int
	Name string
	Active bool
}
