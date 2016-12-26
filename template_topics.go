package main
import "io"
import "strconv"

func init() {
template_topics_handle = template_topics
}

func template_topics(tmpl_topics_vars Page, w io.Writer) {
w.Write([]byte(`<!doctype html>
<html lang="en">
	<head>
		<title>` + tmpl_topics_vars.Title + `</title>
		<link href="/static/main.css" rel="stylesheet" type="text/css">
		<script type="text/javascript" src="/static/jquery-1.12.3.min.js"></script>
		<script type="text/javascript">
		var session = "` + tmpl_topics_vars.CurrentUser.Session + `";
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
if tmpl_topics_vars.CurrentUser.Loggedin {
w.Write([]byte(`
		<li class="menu_left menu_account"><a href="/user/edit/critical/">Account</a></li>
		<li class="menu_left menu_profile"><a href="/user/` + strconv.Itoa(tmpl_topics_vars.CurrentUser.ID) + `">Profile</a></li>
		`))
if tmpl_topics_vars.CurrentUser.Is_Super_Mod {
w.Write([]byte(`<li class="menu_left menu_account"><a href="/panel/">Panel</a></li>`))
}
w.Write([]byte(`
		<li class="menu_left menu_logout"><a href="/accounts/logout?session=` + tmpl_topics_vars.CurrentUser.Session + `">Logout</a></li>
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
if len(tmpl_topics_vars.NoticeList) != 0 {
for _, item := range tmpl_topics_vars.NoticeList {
w.Write([]byte(`<div class="alert">` + item + `</div>`))
}
}
w.Write([]byte(`
<div class="rowblock">
	<div class="rowitem"><a>Topic List</a></div>
</div>
<div class="rowblock">
	`))
if len(tmpl_topics_vars.ItemList) != 0 {
for _, item := range tmpl_topics_vars.ItemList {
w.Write([]byte(`<div class="rowitem passive" style="`))
if item.(TopicUser).Avatar != "" {
w.Write([]byte(`background-image: url(` + item.(TopicUser).Avatar + `);background-position: left;background-repeat: no-repeat;background-size: 64px;padding-left: 72px;`))
}
if item.(TopicUser).Sticky {
w.Write([]byte(`background-color: #FFFFCC;`))
} else {
if item.(TopicUser).Is_Closed {
w.Write([]byte(`background-color: #eaeaea;`))
}
}
w.Write([]byte(`">
		<a href="/topic/` + strconv.Itoa(item.(TopicUser).ID) + `">` + item.(TopicUser).Title + `</a> `))
if item.(TopicUser).Is_Closed {
w.Write([]byte(`<span class="username topic_status_e topic_status_closed" style="float: right;">closed</span>
		`))
} else {
w.Write([]byte(`<span class="username hide_on_micro topic_status_e topic_status_open" style="float: right;">open</span>`))
}
w.Write([]byte(`
		<span class="username hide_on_micro" style="border-right: 0;float: right;">Status</span>
	</div>
	`))
}
} else {
w.Write([]byte(`<div class="rowitem passive">There aren't any topics yet.</div>`))
}
w.Write([]byte(`
</div>
			<!--<link rel="stylesheet" href="https://use.fontawesome.com/8670aa03ca.css">-->
		</div>
	</body>
</html>`))
}
