package main
import "strconv"

func init() {
ctemplates["topic"] = template_topic
}

func template_topic(tmpl_topic_vars Page) (tmpl string) {
var extra_data TopicUser = tmpl_topic_vars.Something.(TopicUser)
tmpl += `<!doctype html>
<html lang="en">
	<head>
		<title>`
tmpl += tmpl_topic_vars.Title
tmpl += `</title>
		<link href="/static/main.css" rel="stylesheet" type="text/css">
		<script type="text/javascript" src="/static/jquery-1.12.3.min.js"></script>
		<script type="text/javascript">
		var session = "`
tmpl += tmpl_topic_vars.CurrentUser.Session
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
if tmpl_topic_vars.CurrentUser.Loggedin {
tmpl += `
		<li class="menu_account"><a href="/user/edit/critical/">Account</a></li>
		<li class="menu_account"><a href="/user/`
tmpl += strconv.Itoa(tmpl_topic_vars.CurrentUser.ID)
tmpl += `">Profile</a></li>
		`
if tmpl_topic_vars.CurrentUser.Is_Super_Mod {
tmpl += `<li class="menu_account"><a href="/panel/forums/">Panel</a></li>`
}
tmpl += `
		<li class="menu_logout"><a href="/accounts/logout?session=`
tmpl += tmpl_topic_vars.CurrentUser.Session
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
if len(tmpl_topic_vars.NoticeList) != 0 {
for _, item := range tmpl_topic_vars.NoticeList {
tmpl += `<div class="alert">`
tmpl += item
tmpl += `</div>`
}
}
tmpl += `
<div class="rowblock">
	<form action='/topic/edit/submit/`
tmpl += strconv.Itoa(extra_data.ID)
tmpl += `' method="post">
		<div class="rowitem"`
if extra_data.Sticky {
tmpl += ` style="background-color: #FFFFEA;"`
}
tmpl += `>
			<a class='topic_name hide_on_edit'>`
tmpl += extra_data.Title
tmpl += `</a> 
			<span class='username topic_status_e topic_status_`
tmpl += extra_data.Status
tmpl += ` hide_on_edit' style="font-weight:normal;float: right;">`
tmpl += extra_data.Status
tmpl += `</span> 
			<span class="username" style="border-right: 0;font-weight: normal;float: right;">Status</span>
			`
if tmpl_topic_vars.CurrentUser.Is_Mod {
tmpl += `
			<a href='/topic/edit/`
tmpl += strconv.Itoa(extra_data.ID)
tmpl += `' class="username hide_on_edit open_edit" style="font-weight: normal;margin-left: 6px;">Edit</a>
			<a href='/topic/delete/submit/`
tmpl += strconv.Itoa(extra_data.ID)
tmpl += `' class="username" style="font-weight: normal;">Delete</a>
			`
if extra_data.Sticky {
tmpl += `<a href='/topic/unstick/submit/`
tmpl += strconv.Itoa(extra_data.ID)
tmpl += `' class="username" style="font-weight: normal;">Unpin</a>`
} else {
tmpl += `<a href='/topic/stick/submit/`
tmpl += strconv.Itoa(extra_data.ID)
tmpl += `' class="username" style="font-weight: normal;">Pin</a>`
}
tmpl += `
			
			<input class='show_on_edit topic_name_input' name="topic_name" value='`
tmpl += extra_data.Title
tmpl += `' type="text" />
			<select name="topic_status" class='show_on_edit topic_status_input' style='float: right;'>
				<option>open</option>
				<option>closed</option>
			</select>
			<button name="topic-button" class="formbutton show_on_edit submit_edit">Update</button>
			`
}
tmpl += `
			<a href="/report/submit/`
tmpl += strconv.Itoa(extra_data.ID)
tmpl += `?session=`
tmpl += tmpl_topic_vars.CurrentUser.Session
tmpl += `&type=topic" class="username report_item" style="font-weight: normal;">Report</a>
		</div>
	</form>
</div>
<div class="rowblock">
	<div class="rowitem passive editable_parent" style="border-bottom: none;`
if extra_data.Avatar != "" {
tmpl += `background-image: url(`
tmpl += extra_data.Avatar
tmpl += `), url(/static/white-dot.jpg);background-position: 0px `
if extra_data.ContentLines <= 5 {
tmpl += `-1`
}
tmpl += `0px;background-repeat: no-repeat, repeat-y;background-size: 128px;padding-left: 136px;`
tmpl += string(extra_data.Css)
}
tmpl += `">
		<span class="hide_on_edit topic_content user_content">`
tmpl += string(tmpl_topic_varsstring(.Something.(TopicUser).Content))
tmpl += `</span>
		<textarea name="topic_content" class="show_on_edit topic_content_input">`
tmpl += string(tmpl_topic_varsstring(.Something.(TopicUser).Content))
tmpl += `</textarea>
		<br /><br />
		<a href="/user/`
tmpl += strconv.Itoa(extra_data.CreatedBy)
tmpl += `" class="username">`
tmpl += extra_data.CreatedByName
tmpl += `</a>
		`
if extra_data.Tag != "" {
tmpl += `<a class="username" style="float: right;">`
tmpl += extra_data.Tag
tmpl += `</a>`
} else {
if extra_data.URLName != "" {
tmpl += `<a href="`
tmpl += extra_data.URL
tmpl += `" class="username" style="color: #505050;float: right;">`
tmpl += extra_data.URLName
tmpl += `</a>
		<a class="username" style="color: #505050;float: right;border-right: 0;">`
tmpl += extra_data.URLPrefix
tmpl += `</a>`
}
}
tmpl += `
	</div>
</div><br />
<div class="rowblock" style="overflow: hidden;">
	`
if len(tmpl_topic_vars.ItemList) != 0 {
for _, item := range tmpl_topic_vars.ItemList {
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
if tmpl_topic_vars.CurrentUser.Is_Mod {
tmpl += `<a href="/reply/edit/submit/`
tmpl += strconv.Itoa(item.(Reply).ID)
tmpl += `"><button class="username edit_item">Edit</button></a>
		<a href="/reply/delete/submit/`
tmpl += strconv.Itoa(item.(Reply).ID)
tmpl += `"><button class="username delete_item">Delete</button></a>`
}
tmpl += `
		<a href="/report/submit/`
tmpl += strconv.Itoa(item.(Reply).ID)
tmpl += `?session=`
tmpl += tmpl_topic_vars.CurrentUser.Session
tmpl += `&type=reply"><button class="username report_item">Report</button></a>
		`
if item.(Reply).Tag != "" {
tmpl += `<a class="username" style="float: right;">`
tmpl += item.(Reply).Tag
tmpl += `</a>`
} else {
if item.(Reply).URLName != "" {
tmpl += `<a href="`
tmpl += item.(Reply).URL
tmpl += `" class="username" style="color: #505050;float: right;" rel="nofollow">`
tmpl += item.(Reply).URLName
tmpl += `</a>
		<a class="username" style="color: #505050;float: right;border-right: 0;">`
tmpl += item.(Reply).URLPrefix
tmpl += `</a>`
}
}
tmpl += `
	</div>`
}
}
tmpl += `
</div>
`
if !tmpl_topic_vars.CurrentUser.Is_Banned {
tmpl += `
<div class="rowblock">
	<form action="/reply/create/" method="post">
		<input name="tid" value='`
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
