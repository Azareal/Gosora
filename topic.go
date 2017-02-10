package main
import "html/template"

type Topic struct
{
	ID int
	Title string
	Content string
	CreatedBy int
	Is_Closed bool
	Sticky bool
	CreatedAt string
	ParentID int
	Status string // Deprecated. Marked for removal.
	IpAddress string
	PostCount int
}

type TopicUser struct
{
	ID int
	Title string
	Content interface{}
	CreatedBy int
	Is_Closed bool
	Sticky bool
	CreatedAt string
	LastReplyAt string
	ParentID int
	Status string // Deprecated. Marked for removal.
	IpAddress string
	PostCount int
	LikeCount int
	
	CreatedByName string
	Avatar string
	Css template.CSS
	ContentLines int
	Tag string
	URL string
	URLPrefix string
	URLName string
	Level int
	Liked bool
}

type TopicsRow struct
{
	ID int
	Title string
	Content interface{}
	CreatedBy int
	Is_Closed bool
	Sticky bool
	CreatedAt string
	LastReplyAt string
	ParentID int
	Status string // Deprecated. Marked for removal.
	IpAddress string
	PostCount int
	LikeCount int
	
	CreatedByName string
	Avatar string
	Css template.CSS
	ContentLines int
	Tag string
	URL string
	URLPrefix string
	URLName string
	Level int
	
	ForumName string //TopicsRow
}
