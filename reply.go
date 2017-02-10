/* Copyright Azareal 2016 - 2017 */
package main
import "html/template"

type Reply struct
{
	ID int
	ParentID int
	Content string
	ContentHtml template.HTML
	CreatedBy int
	CreatedByName string
	CreatedAt string
	LastEdit int
	LastEditBy int
	Avatar string
	Css template.CSS
	ContentLines int
	Tag string
	URL string
	URLPrefix string
	URLName string
	Level int
	IpAddress string
	Liked bool
	LikeCount int
}

