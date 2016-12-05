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
}
