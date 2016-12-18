package main
import "io"
import "strconv"

func init() {
ctemplates["profile"] = template_profile
}

func template_profile(tmpl_profile_vars Page, w io.Writer) {
var extra_data User = tmpl_profile_vars.Something.(User)
w.Write([]byte(`<!doctype html>
<html lang="en">
	<head>
		<title>` + tmpl_profile_vars.Title + `</title>
		<link href="/static/main.css" rel="stylesheet" type="text/css">
		<script type="text/javascript" src="/static/jquery-1.12.3.min.js"></script>
		<script type="text/javascript">
		var session = "` + tmpl_profile_vars.CurrentUser.Session + `";
		</script>
		<script type="text/javascript" src="/static/global.js"></script>
	</head>
	<body>
		<div class="container">
<div class="nav">
	<div class="move_left">
	<div class="move_right">
	<ul>
		<li class="menu_left menu_overview"><a href="/">Overview</a></li>
		<li class="menu_left menu_forums"><a href="/forums/">Forums</a></li>
		<li class="menu_left menu_topics"><a href="/">Topics</a></li>
		<li class="menu_left menu_create_topic"><a href="/topics/create/">Create Topic</a></li>
		`))
if tmpl_profile_vars.CurrentUser.Loggedin {
w.Write([]byte(`
		<li class="menu_left menu_account"><a href="/user/edit/critical/">Account</a></li>
		<li class="menu_left menu_account"><a href="/user/` + strconv.Itoa(tmpl_profile_vars.CurrentUser.ID) + `">Profile</a></li>
		`))
if tmpl_profile_vars.CurrentUser.Is_Super_Mod {
w.Write([]byte(`<li class="menu_left menu_account"><a href="/panel/forums/">Panel</a></li>`))
}
w.Write([]byte(`
		<li class="menu_left menu_logout"><a href="/accounts/logout?session=` + tmpl_profile_vars.CurrentUser.Session + `">Logout</a></li>
		`))
} else {
w.Write([]byte(`
		<li class="menu_left menu_register"><a href="/accounts/create/">Register</a></li>
		<li class="menu_left menu_login"><a href="/accounts/login/">Login</a></li>
		`))
}
w.Write([]byte(`
	</ul>
	</div>
	</div>
	<div style="clear: both;"></div>
</div>
`))
if len(tmpl_profile_vars.NoticeList) != 0 {
for _, item := range tmpl_profile_vars.NoticeList {
w.Write([]byte(`<div class="alert">` + item + `</div>`))
}
}
w.Write([]byte(`
<div class="colblock_left" style="max-width: 220px;">
	<div class="rowitem" style="padding: 0;"><img src="` + extra_data.Avatar + `" style="max-width: 100%;margin: 0;"/></div>
	<div class="rowitem" style="text-transform: capitalize;">
	<span style="font-size: 18px;">` + extra_data.Name + `</span>`))
if extra_data.Tag != "" {
w.Write([]byte(`<span class="username" style="float: right;">` + extra_data.Tag + `</span>`))
}
w.Write([]byte(`
	</div>
	<div class="rowitem passive">
		<a class="username">Add Friend</a>
		`))
if tmpl_profile_vars.CurrentUser.Is_Super_Mod && !extra_data.Is_Super_Mod {
w.Write([]byte(`
		`))
if extra_data.Is_Banned {
w.Write([]byte(`<a href="/users/unban/` + strconv.Itoa(extra_data.ID) + `" class="username">Unban</a>`))
} else {
w.Write([]byte(`<a href="/users/ban/` + strconv.Itoa(extra_data.ID) + `" class="username">Ban</a>`))
}
w.Write([]byte(`
		`))
}
w.Write([]byte(`
		<a href="/report/submit/` + strconv.Itoa(extra_data.ID) + `?session=` + tmpl_profile_vars.CurrentUser.Session + `&type=user" class="username report_item">Report</a>
	</div>
</div>
<div class="colblock_right">
	<div class="rowitem"><a>Comments</a></div>
</div>
<div class="colblock_right" style="overflow: hidden;">
	`))
if len(tmpl_profile_vars.ItemList) != 0 {
for _, item := range tmpl_profile_vars.ItemList {
w.Write([]byte(`
	<div class="rowitem passive deletable_block editable_parent" style="`))
if item.(Reply).Avatar != "" {
w.Write([]byte(`background-image: url(` + item.(Reply).Avatar + `), url(/static/white-dot.jpg);background-position: 0px `))
if item.(Reply).ContentLines <= 5 {
w.Write([]byte(`-1`))
}
w.Write([]byte(`0px;background-repeat: no-repeat, repeat-y;background-size: 128px;padding-left: 136px;` + string(item.(Reply).Css)))
}
w.Write([]byte(`">
		<span class="editable_block user_content">` + string(item.(Reply).ContentHtml) + `</span>
		<br /><br />
		<a href="/user/` + strconv.Itoa(item.(Reply).CreatedBy) + `" class="username">` + item.(Reply).CreatedByName + `</a>
		`))
if tmpl_profile_vars.CurrentUser.Is_Mod {
w.Write([]byte(`<a href="/profile/reply/edit/submit/` + strconv.Itoa(item.(Reply).ID) + `"><button class="username edit_item">Edit</button></a>
		<a href="/profile/reply/delete/submit/` + strconv.Itoa(item.(Reply).ID) + `"><button class="username delete_item">Delete</button></a>`))
}
w.Write([]byte(`
		<a href="/report/submit/` + strconv.Itoa(item.(Reply).ID) + `?session=` + tmpl_profile_vars.CurrentUser.Session + `&type=user-reply"><button class="username report_item">Report</button></a>
		`))
if item.(Reply).Tag != "" {
w.Write([]byte(`<a class="username" style="float: right;">` + item.(Reply).Tag + `</a>`))
}
w.Write([]byte(`
	</div>`))
}
}
w.Write([]byte(`
</div>
`))
if !tmpl_profile_vars.CurrentUser.Is_Banned {
w.Write([]byte(`
<div class="colblock_right">
	<form action="/profile/reply/create/" method="post">
		<input name="uid" value='` + strconv.Itoa(extra_data.ID) + `' type="hidden" />
		<div class="formrow">
			<div class="formitem"><textarea name="reply-content" placeholder="Insert reply here"></textarea></div>
		</div>
		<div class="formrow">
			<div class="formitem"><button name="reply-button" class="formbutton">Create Reply</button></div>
		</div>
	</form>
</div>
`))
}
w.Write([]byte(`
			<!--<link rel="stylesheet" href="https://use.fontawesome.com/8670aa03ca.css">-->
		</div>
	</body>
</html>`))
}
