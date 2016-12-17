package main
import "io"
import "strconv"
import "html/template"

func init() {
ctemplates["topic"] = template_topic
}

func template_topic(tmpl_topic_vars Page, w io.Writer) {
var extra_data TopicUser = tmpl_topic_vars.Something.(TopicUser)
w.Write([]byte(`<!doctype html>
<html lang="en">
	<head>
		<title>` + tmpl_topic_vars.Title + `</title>
		<link href="/static/main.css" rel="stylesheet" type="text/css">
		<script type="text/javascript" src="/static/jquery-1.12.3.min.js"></script>
		<script type="text/javascript">
		var session = "` + tmpl_topic_vars.CurrentUser.Session + `";
		</script>
		<script type="text/javascript" src="/static/global.js"></script>
	</head>
	<body>
		<div class="container">
<div class="nav">
	<ul>
		<li class="menu_overview"><a href="/">Overview</a></li>
		<li class="menu_forums"><a href="/forums/">Forums</a></li>
		<li class="menu_topics"><a href="/">Topics</a></li>
		<li class="menu_create_topic"><a href="/topics/create/">Create Topic</a></li>
		`))
if tmpl_topic_vars.CurrentUser.Loggedin {
w.Write([]byte(`
		<li class="menu_account"><a href="/user/edit/critical/">Account</a></li>
		<li class="menu_account"><a href="/user/` + strconv.Itoa(tmpl_topic_vars.CurrentUser.ID) + `">Profile</a></li>
		`))
if tmpl_topic_vars.CurrentUser.Is_Super_Mod {
w.Write([]byte(`<li class="menu_account"><a href="/panel/forums/">Panel</a></li>`))
}
w.Write([]byte(`
		<li class="menu_logout"><a href="/accounts/logout?session=` + tmpl_topic_vars.CurrentUser.Session + `">Logout</a></li>
		`))
} else {
w.Write([]byte(`
		<li class="menu_register"><a href="/accounts/create/">Register</a></li>
		<li class="menu_login"><a href="/accounts/login/">Login</a></li>
		`))
}
w.Write([]byte(`
	</ul>
</div>
`))
if len(tmpl_topic_vars.NoticeList) != 0 {
for _, item := range tmpl_topic_vars.NoticeList {
w.Write([]byte(`<div class="alert">` + item + `</div>`))
}
}
w.Write([]byte(`
<div class="rowblock">
	<form action='/topic/edit/submit/` + strconv.Itoa(extra_data.ID) + `' method="post">
		<div class="rowitem"`))
if extra_data.Sticky {
w.Write([]byte(` style="background-color: #FFFFEA;"`))
}
w.Write([]byte(`>
			<a class='topic_name hide_on_edit'>` + extra_data.Title + `</a> 
			<span class='username topic_status_e topic_status_` + extra_data.Status + ` hide_on_edit' style="font-weight:normal;float: right;">` + extra_data.Status + `</span> 
			<span class="username" style="border-right: 0;font-weight: normal;float: right;">Status</span>
			`))
if tmpl_topic_vars.CurrentUser.Is_Mod {
w.Write([]byte(`
			<a href='/topic/edit/` + strconv.Itoa(extra_data.ID) + `' class="username hide_on_edit open_edit" style="font-weight: normal;margin-left: 6px;">Edit</a>
			<a href='/topic/delete/submit/` + strconv.Itoa(extra_data.ID) + `' class="username" style="font-weight: normal;">Delete</a>
			`))
if extra_data.Sticky {
w.Write([]byte(`<a href='/topic/unstick/submit/` + strconv.Itoa(extra_data.ID) + `' class="username" style="font-weight: normal;">Unpin</a>`))
} else {
w.Write([]byte(`<a href='/topic/stick/submit/` + strconv.Itoa(extra_data.ID) + `' class="username" style="font-weight: normal;">Pin</a>`))
}
w.Write([]byte(`
			
			<input class='show_on_edit topic_name_input' name="topic_name" value='` + extra_data.Title + `' type="text" />
			<select name="topic_status" class='show_on_edit topic_status_input' style='float: right;'>
				<option>open</option>
				<option>closed</option>
			</select>
			<button name="topic-button" class="formbutton show_on_edit submit_edit">Update</button>
			`))
}
w.Write([]byte(`
			<a href="/report/submit/` + strconv.Itoa(extra_data.ID) + `?session=` + tmpl_topic_vars.CurrentUser.Session + `&type=topic" class="username report_item" style="font-weight: normal;">Report</a>
		</div>
	</form>
</div>
<div class="rowblock">
	<div class="rowitem passive editable_parent" style="border-bottom: none;`))
if extra_data.Avatar != "" {
w.Write([]byte(`background-image: url(` + extra_data.Avatar + `), url(/static/white-dot.jpg);background-position: 0px `))
if extra_data.ContentLines <= 5 {
w.Write([]byte(`-1`))
}
w.Write([]byte(`0px;background-repeat: no-repeat, repeat-y;background-size: 128px;padding-left: 136px;` + string(extra_data.Css)))
}
w.Write([]byte(`">
		<span class="hide_on_edit topic_content user_content">` + string(extra_data.Content.(template.HTML)) + `</span>
		<textarea name="topic_content" class="show_on_edit topic_content_input">` + string(extra_data.Content.(template.HTML)) + `</textarea>
		<br /><br />
		<a href="/user/` + strconv.Itoa(extra_data.CreatedBy) + `" class="username">` + extra_data.CreatedByName + `</a>
		`))
if extra_data.Tag != "" {
w.Write([]byte(`<a class="username" style="float: right;">` + extra_data.Tag + `</a>`))
} else {
if extra_data.URLName != "" {
w.Write([]byte(`<a href="` + extra_data.URL + `" class="username" style="color: #505050;float: right;">` + extra_data.URLName + `</a>
		<a class="username" style="color: #505050;float: right;border-right: 0;">` + extra_data.URLPrefix + `</a>`))
}
}
w.Write([]byte(`
	</div>
</div><br />
<div class="rowblock" style="overflow: hidden;">
	`))
if len(tmpl_topic_vars.ItemList) != 0 {
for _, item := range tmpl_topic_vars.ItemList {
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
if tmpl_topic_vars.CurrentUser.Is_Mod {
w.Write([]byte(`<a href="/reply/edit/submit/` + strconv.Itoa(item.(Reply).ID) + `"><button class="username edit_item">Edit</button></a>
		<a href="/reply/delete/submit/` + strconv.Itoa(item.(Reply).ID) + `"><button class="username delete_item">Delete</button></a>`))
}
w.Write([]byte(`
		<a href="/report/submit/` + strconv.Itoa(item.(Reply).ID) + `?session=` + tmpl_topic_vars.CurrentUser.Session + `&type=reply"><button class="username report_item">Report</button></a>
		`))
if item.(Reply).Tag != "" {
w.Write([]byte(`<a class="username" style="float: right;">` + item.(Reply).Tag + `</a>`))
} else {
if item.(Reply).URLName != "" {
w.Write([]byte(`<a href="` + item.(Reply).URL + `" class="username" style="color: #505050;float: right;" rel="nofollow">` + item.(Reply).URLName + `</a>
		<a class="username" style="color: #505050;float: right;border-right: 0;">` + item.(Reply).URLPrefix + `</a>`))
}
}
w.Write([]byte(`
	</div>`))
}
}
w.Write([]byte(`
</div>
`))
if !tmpl_topic_vars.CurrentUser.Is_Banned {
w.Write([]byte(`
<div class="rowblock">
	<form action="/reply/create/" method="post">
		<input name="tid" value='` + strconv.Itoa(extra_data.ID) + `' type="hidden" />
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
