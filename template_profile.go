package main
import "strconv"

func init() {
ctemplates["profile"] = template_profile
}

func template_profile(tmpl_profile_vars Page) (tmpl string) {
var extra_data User = tmpl_profile_vars.Something.(User)
tmpl += `<!doctype html>
<html lang="en">
	<head>
		<title>`
tmpl += tmpl_profile_vars.Title
tmpl += `</title>
		<link href="/static/main.css" rel="stylesheet" type="text/css">
		<script type="text/javascript" src="/static/jquery-1.12.3.min.js"></script>
		<script type="text/javascript">
		var session = "`
tmpl += tmpl_profile_vars.CurrentUser.Session
tmpl += `";
		</script>
		<script type="text/javascript" src="/static/global.js"></script>
	</head>
	<body>
		<div class="container">
`
tmpl += `<div class="nav">
	<ul>
		<li class="menu_overview"><a href="/">Overview</a></li>
		<li class="menu_forums"><a href="/forums/">Forums</a></li>
		<li class="menu_topics"><a href="/">Topics</a></li>
		<li class="menu_create_topic"><a href="/topics/create/">Create Topic</a></li>
		`
if tmpl_profile_vars.CurrentUser.Loggedin {
tmpl += `
		<li class="menu_account"><a href="/user/edit/critical/">Account</a></li>
		<li class="menu_account"><a href="/user/`
tmpl += strconv.Itoa(tmpl_profile_vars.CurrentUser.ID)
tmpl += `">Profile</a></li>
		`
if tmpl_profile_vars.CurrentUser.Is_Super_Mod {
tmpl += `<li class="menu_account"><a href="/panel/forums/">Panel</a></li>`
}
tmpl += `
		<li class="menu_logout"><a href="/accounts/logout?session=`
tmpl += tmpl_profile_vars.CurrentUser.Session
tmpl += `">Logout</a></li>
		`
} else {
tmpl += `
		<li class="menu_register"><a href="/accounts/create/">Register</a></li>
		<li class="menu_login"><a href="/accounts/login/">Login</a></li>
		`
}
tmpl += `
	</ul>
</div>`
tmpl += `
`
if len(tmpl_profile_vars.NoticeList) != 0 {
for _, item := range tmpl_profile_vars.NoticeList {
tmpl += `<div class="alert">`
tmpl += item
tmpl += `</div>`
}
}
tmpl += `
<div class="colblock_left" style="max-width: 220px;">
	<div class="rowitem" style="padding: 0;"><img src="`
tmpl += extra_data.Avatar
tmpl += `" style="max-width: 100%;margin: 0;"/></div>
	<div class="rowitem" style="text-transform: capitalize;">
	<span style="font-size: 18px;">`
tmpl += extra_data.Name
tmpl += `</span>`
if extra_data.Tag != "" {
tmpl += `<span class="username" style="float: right;">`
tmpl += extra_data.Tag
tmpl += `</span>`
}
tmpl += `
	</div>
	<div class="rowitem passive">
		<a class="username">Add Friend</a>
		`
if tmpl_profile_vars.CurrentUser.Is_Super_Mod && !extra_data.Is_Super_Mod {
tmpl += `
		`
if extra_data.Is_Banned {
tmpl += `<a href="/users/unban/`
tmpl += strconv.Itoa(extra_data.ID)
tmpl += `" class="username">Unban</a>`
} else {
tmpl += `<a href="/users/ban/`
tmpl += strconv.Itoa(extra_data.ID)
tmpl += `" class="username">Ban</a>`
}
tmpl += `
		`
}
tmpl += `
		<a href="/report/submit/`
tmpl += strconv.Itoa(extra_data.ID)
tmpl += `?session=`
tmpl += tmpl_profile_vars.CurrentUser.Session
tmpl += `&type=user" class="username report_item">Report</a>
	</div>
</div>
<div class="colblock_right">
	<div class="rowitem"><a>Comments</a></div>
</div>
<div class="colblock_right" style="overflow: hidden;">
	`
if len(tmpl_profile_vars.ItemList) != 0 {
for _, item := range tmpl_profile_vars.ItemList {
tmpl += `
	<div class="rowitem passive deletable_block editable_parent" style="`
if item.(Reply).Avatar != "" {
tmpl += `background-image: url(`
tmpl += item.(Reply).Avatar
tmpl += `), url(/static/white-dot.jpg);background-position: 0px `
if item.(Reply).ContentLines <= 5 {
tmpl += `-1`
}
tmpl += `0px;background-repeat: no-repeat, repeat-y;background-size: 128px;padding-left: 136px;`
tmpl += string(item.(Reply).Css)
}
tmpl += `">
		<span class="editable_block user_content">`
tmpl += string(item.(Reply).ContentHtml)
tmpl += `</span>
		<br /><br />
		<a href="/user/`
tmpl += strconv.Itoa(item.(Reply).CreatedBy)
tmpl += `" class="username">`
tmpl += item.(Reply).CreatedByName
tmpl += `</a>
		`
if tmpl_profile_vars.CurrentUser.Is_Mod {
tmpl += `<a href="/profile/reply/edit/submit/`
tmpl += strconv.Itoa(item.(Reply).ID)
tmpl += `"><button class="username edit_item">Edit</button></a>
		<a href="/profile/reply/delete/submit/`
tmpl += strconv.Itoa(item.(Reply).ID)
tmpl += `"><button class="username delete_item">Delete</button></a>`
}
tmpl += `
		<a href="/report/submit/`
tmpl += strconv.Itoa(item.(Reply).ID)
tmpl += `?session=`
tmpl += tmpl_profile_vars.CurrentUser.Session
tmpl += `&type=user-reply"><button class="username report_item">Report</button></a>
		`
if item.(Reply).Tag != "" {
tmpl += `<a class="username" style="float: right;">`
tmpl += item.(Reply).Tag
tmpl += `</a>`
}
tmpl += `
	</div>`
}
}
tmpl += `
</div>
`
if !tmpl_profile_vars.CurrentUser.Is_Banned {
tmpl += `
<div class="colblock_right">
	<form action="/profile/reply/create/" method="post">
		<input name="uid" value='`
tmpl += strconv.Itoa(extra_data.ID)
tmpl += `' type="hidden" />
		<div class="formrow">
			<div class="formitem"><textarea name="reply-content" placeholder="Insert reply here"></textarea></div>
		</div>
		<div class="formrow">
			<div class="formitem"><button name="reply-button" class="formbutton">Create Reply</button></div>
		</div>
	</form>
</div>
`
}
tmpl += `
`
tmpl += `			<!--<link rel="stylesheet" href="https://use.fontawesome.com/8670aa03ca.css">-->
		</div>
	</body>
</html>`
return tmpl
}
