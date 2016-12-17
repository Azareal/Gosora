package main
import "io"
import "strconv"
import "html/template"

/*func init() {
ctemplates["topic"] = template_topic
}*/

func template_topic2(tmpl_topic_vars Page, w io.Writer) {
var extra_data TopicUser = tmpl_topic_vars.Something.(TopicUser)
w.Write([]byte(`<!doctype html>
<html lang="en">
	<head>
		<title>`))
w.Write([]byte(tmpl_topic_vars.Title))
w.Write([]byte(`</title>
		<link href="/static/main.css" rel="stylesheet" type="text/css">
		<script type="text/javascript" src="/static/jquery-1.12.3.min.js"></script>
		<script type="text/javascript">
		var session = "`))
w.Write([]byte(tmpl_topic_vars.CurrentUser.Session))
w.Write([]byte(`";
		</script>
		<script type="text/javascript" src="/static/global.js"></script>
	</head>
	<body>
		<div class="container">
`))
w.Write([]byte(`<div class="nav">
	<ul>
		<li class="menu_overview"><a href="/">Overview</a></li>
		<li class="menu_forums"><a href="/forums/">Forums</a></li>
		<li class="menu_topics"><a href="/">Topics</a></li>
		<li class="menu_create_topic"><a href="/topics/create/">Create Topic</a></li>
		`))
if tmpl_topic_vars.CurrentUser.Loggedin {
w.Write([]byte(`
		<li class="menu_account"><a href="/user/edit/critical/">Account</a></li>
		<li class="menu_account"><a href="/user/`))
w.Write([]byte(strconv.Itoa(tmpl_topic_vars.CurrentUser.ID)))
w.Write([]byte(`">Profile</a></li>
		`))
if tmpl_topic_vars.CurrentUser.Is_Super_Mod {
w.Write([]byte(`<li class="menu_account"><a href="/panel/forums/">Panel</a></li>`))
}
w.Write([]byte(`
		<li class="menu_logout"><a href="/accounts/logout?session=`))
w.Write([]byte(tmpl_topic_vars.CurrentUser.Session))
w.Write([]byte(`">Logout</a></li>
		`))
} else {
w.Write([]byte(`
		<li class="menu_register"><a href="/accounts/create/">Register</a></li>
		<li class="menu_login"><a href="/accounts/login/">Login</a></li>
		`))
}
w.Write([]byte(`
	</ul>
</div>`))
w.Write([]byte(`
`))
if len(tmpl_topic_vars.NoticeList) != 0 {
for _, item := range tmpl_topic_vars.NoticeList {
w.Write([]byte(`<div class="alert">`))
w.Write([]byte(item))
w.Write([]byte(`</div>`))
}
}
w.Write([]byte(`
<div class="rowblock">
	<form action='/topic/edit/submit/`))
w.Write([]byte(strconv.Itoa(extra_data.ID)))
w.Write([]byte(`' method="post">
		<div class="rowitem"`))
if extra_data.Sticky {
w.Write([]byte(` style="background-color: #FFFFEA;"`))
}
w.Write([]byte(`>
			<a class='topic_name hide_on_edit'>`))
w.Write([]byte(extra_data.Title))
w.Write([]byte(`</a> 
			<span class='username topic_status_e topic_status_`))
w.Write([]byte(extra_data.Status))
w.Write([]byte(` hide_on_edit' style="font-weight:normal;float: right;">`))
w.Write([]byte(extra_data.Status))
w.Write([]byte(`</span> 
			<span class="username" style="border-right: 0;font-weight: normal;float: right;">Status</span>
			`))
if tmpl_topic_vars.CurrentUser.Is_Mod {
w.Write([]byte(`
			<a href='/topic/edit/`))
w.Write([]byte(strconv.Itoa(extra_data.ID)))
w.Write([]byte(`' class="username hide_on_edit open_edit" style="font-weight: normal;margin-left: 6px;">Edit</a>
			<a href='/topic/delete/submit/`))
w.Write([]byte(strconv.Itoa(extra_data.ID)))
w.Write([]byte(`' class="username" style="font-weight: normal;">Delete</a>
			`))
if extra_data.Sticky {
w.Write([]byte(`<a href='/topic/unstick/submit/`))
w.Write([]byte(strconv.Itoa(extra_data.ID)))
w.Write([]byte(`' class="username" style="font-weight: normal;">Unpin</a>`))
} else {
w.Write([]byte(`<a href='/topic/stick/submit/`))
w.Write([]byte(strconv.Itoa(extra_data.ID)))
w.Write([]byte(`' class="username" style="font-weight: normal;">Pin</a>`))
}
w.Write([]byte(`
			
			<input class='show_on_edit topic_name_input' name="topic_name" value='`))
w.Write([]byte(extra_data.Title))
w.Write([]byte(`' type="text" />
			<select name="topic_status" class='show_on_edit topic_status_input' style='float: right;'>
				<option>open</option>
				<option>closed</option>
			</select>
			<button name="topic-button" class="formbutton show_on_edit submit_edit">Update</button>
			`))
}
w.Write([]byte(`
			<a href="/report/submit/`))
w.Write([]byte(strconv.Itoa(extra_data.ID)))
w.Write([]byte(`?session=`))
w.Write([]byte(tmpl_topic_vars.CurrentUser.Session))
w.Write([]byte(`&type=topic" class="username report_item" style="font-weight: normal;">Report</a>
		</div>
	</form>
</div>
<div class="rowblock">
	<div class="rowitem passive editable_parent" style="border-bottom: none;`))
if extra_data.Avatar != "" {
w.Write([]byte(`background-image: url(`))
w.Write([]byte(extra_data.Avatar))
w.Write([]byte(`), url(/static/white-dot.jpg);background-position: 0px `))
if extra_data.ContentLines <= 5 {
w.Write([]byte(`-1`))
}
w.Write([]byte(`0px;background-repeat: no-repeat, repeat-y;background-size: 128px;padding-left: 136px;`))
w.Write([]byte(string(extra_data.Css)))
}
w.Write([]byte(`">
		<span class="hide_on_edit topic_content user_content">`))
w.Write([]byte(string(extra_data.Content.(template.HTML))))
w.Write([]byte(`</span>
		<textarea name="topic_content" class="show_on_edit topic_content_input">`))
w.Write([]byte(string(extra_data.Content.(template.HTML))))
w.Write([]byte(`</textarea>
		<br /><br />
		<a href="/user/`))
w.Write([]byte(strconv.Itoa(extra_data.CreatedBy)))
w.Write([]byte(`" class="username">`))
w.Write([]byte(extra_data.CreatedByName))
w.Write([]byte(`</a>
		`))
if extra_data.Tag != "" {
w.Write([]byte(`<a class="username" style="float: right;">`))
w.Write([]byte(extra_data.Tag))
w.Write([]byte(`</a>`))
} else {
if extra_data.URLName != "" {
w.Write([]byte(`<a href="`))
w.Write([]byte(extra_data.URL))
w.Write([]byte(`" class="username" style="color: #505050;float: right;">`))
w.Write([]byte(extra_data.URLName))
w.Write([]byte(`</a>
		<a class="username" style="color: #505050;float: right;border-right: 0;">`))
w.Write([]byte(extra_data.URLPrefix))
w.Write([]byte(`</a>`))
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
w.Write([]byte(`background-image: url(`))
w.Write([]byte(item.(Reply).Avatar))
w.Write([]byte(`), url(/static/white-dot.jpg);background-position: 0px `))
if item.(Reply).ContentLines <= 5 {
w.Write([]byte(`-1`))
}
w.Write([]byte(`0px;background-repeat: no-repeat, repeat-y;background-size: 128px;padding-left: 136px;`))
w.Write([]byte(string(item.(Reply).Css)))
}
w.Write([]byte(`">
		<span class="editable_block user_content">`))
w.Write([]byte(string(item.(Reply).ContentHtml)))
w.Write([]byte(`</span>
		<br /><br />
		<a href="/user/`))
w.Write([]byte(strconv.Itoa(item.(Reply).CreatedBy)))
w.Write([]byte(`" class="username">`))
w.Write([]byte(item.(Reply).CreatedByName))
w.Write([]byte(`</a>
		`))
if tmpl_topic_vars.CurrentUser.Is_Mod {
w.Write([]byte(`<a href="/reply/edit/submit/`))
w.Write([]byte(strconv.Itoa(item.(Reply).ID)))
w.Write([]byte(`"><button class="username edit_item">Edit</button></a>
		<a href="/reply/delete/submit/`))
w.Write([]byte(strconv.Itoa(item.(Reply).ID)))
w.Write([]byte(`"><button class="username delete_item">Delete</button></a>`))
}
w.Write([]byte(`
		<a href="/report/submit/`))
w.Write([]byte(strconv.Itoa(item.(Reply).ID)))
w.Write([]byte(`?session=`))
w.Write([]byte(tmpl_topic_vars.CurrentUser.Session))
w.Write([]byte(`&type=reply"><button class="username report_item">Report</button></a>
		`))
if item.(Reply).Tag != "" {
w.Write([]byte(`<a class="username" style="float: right;">`))
w.Write([]byte(item.(Reply).Tag))
w.Write([]byte(`</a>`))
} else {
if item.(Reply).URLName != "" {
w.Write([]byte(`<a href="`))
w.Write([]byte(item.(Reply).URL))
w.Write([]byte(`" class="username" style="color: #505050;float: right;" rel="nofollow">`))
w.Write([]byte(item.(Reply).URLName))
w.Write([]byte(`</a>
		<a class="username" style="color: #505050;float: right;border-right: 0;">`))
w.Write([]byte(item.(Reply).URLPrefix))
w.Write([]byte(`</a>`))
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
		<input name="tid" value='`))
w.Write([]byte(strconv.Itoa(extra_data.ID)))
w.Write([]byte(`' type="hidden" />
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
`))
w.Write([]byte(`			<!--<link rel="stylesheet" href="https://use.fontawesome.com/8670aa03ca.css">-->
		</div>
	</body>
</html>`))
}
