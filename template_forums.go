package main
import "io"
import "strconv"

func init() {
template_forums_handle = template_forums
}

func template_forums(tmpl_forums_vars Page, w io.Writer) {
w.Write([]byte(`<!doctype html>
<html lang="en">
	<head>
		<title>` + tmpl_forums_vars.Title + `</title>
		<link href="/static/main.css" rel="stylesheet" type="text/css">
		<script type="text/javascript" src="/static/jquery-1.12.3.min.js"></script>
		<script type="text/javascript">
		var session = "` + tmpl_forums_vars.CurrentUser.Session + `";
		</script>
		<script type="text/javascript" src="/static/global.js"></script>
		<meta name="viewport" content="width=device-width,initial-scale = 1.0, maximum-scale=1.0,user-scalable=no" />
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
if tmpl_forums_vars.CurrentUser.Loggedin {
w.Write([]byte(`
		<li class="menu_left menu_account"><a href="/user/edit/critical/">Account</a></li>
		<li class="menu_left menu_profile"><a href="/user/` + strconv.Itoa(tmpl_forums_vars.CurrentUser.ID) + `">Profile</a></li>
		`))
if tmpl_forums_vars.CurrentUser.Is_Super_Mod {
w.Write([]byte(`<li class="menu_left menu_account"><a href="/panel/">Panel</a></li>`))
}
w.Write([]byte(`
		<li class="menu_left menu_logout"><a href="/accounts/logout?session=` + tmpl_forums_vars.CurrentUser.Session + `">Logout</a></li>
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
if len(tmpl_forums_vars.NoticeList) != 0 {
for _, item := range tmpl_forums_vars.NoticeList {
w.Write([]byte(`<div class="alert">` + item + `</div>`))
}
}
w.Write([]byte(`
<div class="rowblock">
	`))
if len(tmpl_forums_vars.ItemList) != 0 {
for _, item := range tmpl_forums_vars.ItemList {
w.Write([]byte(`<div class="rowitem">
		<a href="/forum/` + strconv.Itoa(item.(Forum).ID) + `" style="font-size: 20px;position:relative;top: -2px;font-weight: normal;text-transform: none;">` + item.(Forum).Name + `</a>
		<a href="/topic/` + strconv.Itoa(item.(Forum).LastTopicID) + `" style="font-weight: normal;text-transform: none;float: right;">` + item.(Forum).LastTopic + ` <small style="font-size: 12px;">` + item.(Forum).LastTopicTime + `</small></a>
	</div>
	`))
}
} else {
w.Write([]byte(`<div class="rowitem passive">You don't have access to any forums.</div>`))
}
w.Write([]byte(`
</div>
			<!--<link rel="stylesheet" href="https://use.fontawesome.com/8670aa03ca.css">-->
		</div>
	</body>
</html>`))
}
