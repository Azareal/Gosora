package main

var header_0 []byte = []byte(`<!doctype html>
<html lang="en">
	<head>
		<title>`)
var header_1 []byte = []byte(`</title>
		<link href="/static/main.css" rel="stylesheet" type="text/css">
		<script type="text/javascript" src="/static/jquery-3.1.1.min.js"></script>
		<script type="text/javascript">var session = "`)
var header_2 []byte = []byte(`";
		</script>
		<script type="text/javascript" src="/static/global.js"></script>
		<meta name="viewport" content="width=device-width,initial-scale = 1.0, maximum-scale=1.0,user-scalable=no" />
	</head>
	<body>
		<div class="container">
`)
var menu_0 []byte = []byte(`<div class="nav">
	<div class="move_left">
	<div class="move_right">
	<ul>
		<li class="menu_left menu_overview"><a href="/">Overview</a></li>
		<li class="menu_left menu_forums"><a href="/forums/">Forums</a></li>
		<li class="menu_left menu_topics"><a href="/">Topics</a></li>
		<li class="menu_left menu_create_topic"><a href="/topics/create/">Create Topic</a></li>
		`)
var menu_1 []byte = []byte(`
		<li class="menu_left menu_account"><a href="/user/edit/critical/">Account</a></li>
		<li class="menu_left menu_profile"><a href="/user/`)
var menu_2 []byte = []byte(`">Profile</a></li>
		`)
var menu_3 []byte = []byte(`<li class="menu_left menu_account"><a href="/panel/">Panel</a></li>`)
var menu_4 []byte = []byte(`
		<li class="menu_left menu_logout"><a href="/accounts/logout?session=`)
var menu_5 []byte = []byte(`">Logout</a></li>
		`)
var menu_6 []byte = []byte(`
		<li class="menu_left menu_register"><a href="/accounts/create/">Register</a></li>
		<li class="menu_left menu_login"><a href="/accounts/login/">Login</a></li>
		`)
var menu_7 []byte = []byte(`
	</ul>
	</div>
	</div>
	<div style="clear: both;"></div>
</div>`)
var header_3 []byte = []byte(`
<div id="back"><div id="main">
`)
var header_4 []byte = []byte(`<div class="alert">`)
var header_5 []byte = []byte(`</div>`)
var topic_0 []byte = []byte(`<div id="prevFloat" class="prev_button"><a class="prev_link" href="/topic/`)
var topic_1 []byte = []byte(`?page=`)
var topic_2 []byte = []byte(`">&lt;</a></div>`)
var topic_3 []byte = []byte(`<link rel="prerender" href="/topic/`)
var topic_4 []byte = []byte(`?page=`)
var topic_5 []byte = []byte(`" />
<div id="nextFloat" class="next_button"><a class="next_link" href="/topic/`)
var topic_6 []byte = []byte(`?page=`)
var topic_7 []byte = []byte(`">&gt;</a></div>`)
var topic_8 []byte = []byte(`
<div class="rowblock topic_block">
	<form action='/topic/edit/submit/`)
var topic_9 []byte = []byte(`' method="post">
		<div class="rowitem topic_item"`)
var topic_10 []byte = []byte(` style="background-color:#FFFFEA;"`)
var topic_11 []byte = []byte(` style="background-color:#eaeaea;"`)
var topic_12 []byte = []byte(`>
			<a class='topic_name hide_on_edit'>`)
var topic_13 []byte = []byte(`</a> 
			`)
var topic_14 []byte = []byte(`<span class='username hide_on_micro topic_status_e topic_status_closed hide_on_edit' title='Status: Closed' style="font-weight:normal;float: right;position:relative;top:-5px;">&#x1F512;&#xFE0E</span>`)
var topic_15 []byte = []byte(`
			<input class='show_on_edit topic_name_input' name="topic_name" value='`)
var topic_16 []byte = []byte(`' type="text" />
			`)
var topic_17 []byte = []byte(`<select name="topic_status" class='show_on_edit topic_status_input' style='float: right;'>
				<option>open</option>
				<option>closed</option>
			</select>`)
var topic_18 []byte = []byte(`
			<button name="topic-button" class="formbutton show_on_edit submit_edit">Update</button>
			`)
var topic_19 []byte = []byte(`
		</div>
	</form>
</div>
<div class="rowblock post_container">
	<div class="rowitem passive editable_parent post_item" style="border-bottom: none;`)
var topic_20 []byte = []byte(`background-image:url(`)
var topic_21 []byte = []byte(`), url(/static/white-dot.jpg);background-position: 0px `)
var topic_22 []byte = []byte(`-1`)
var topic_23 []byte = []byte(`0px;background-repeat:no-repeat, repeat-y;background-size:128px;padding-left:136px;`)
var topic_24 []byte = []byte(`">
		<p class="hide_on_edit topic_content user_content" style="margin:0;padding:0;">`)
var topic_25 []byte = []byte(`</p>
		<textarea name="topic_content" class="show_on_edit topic_content_input">`)
var topic_26 []byte = []byte(`</textarea><br /><br />
		<a href="/user/`)
var topic_27 []byte = []byte(`" class="username real_username">`)
var topic_28 []byte = []byte(`</a>&nbsp;
		`)
var topic_29 []byte = []byte(`<a href="/topic/like/submit/`)
var topic_30 []byte = []byte(`" class="mod_button" title="Love it" style="color:#202020;"><button class="username" style="`)
var topic_31 []byte = []byte(`background-color:/*#eaffea*/#D6FFD6;`)
var topic_32 []byte = []byte(`">😀</button></a>&nbsp;`)
var topic_33 []byte = []byte(`<a href='/topic/edit/`)
var topic_34 []byte = []byte(`' class="mod_button open_edit" style="font-weight:normal;" title="Edit Topic"><button class="username">🖊️</button></a>&nbsp;`)
var topic_35 []byte = []byte(`<a href='/topic/delete/submit/`)
var topic_36 []byte = []byte(`' class="mod_button" style="font-weight:normal;" title="Delete Topic"><button class="username">🗑️</button></a>&nbsp;`)
var topic_37 []byte = []byte(`<a class="mod_button" href='/topic/unstick/submit/`)
var topic_38 []byte = []byte(`' style="font-weight:normal;" title="Unpin Topic"><button class="username" style="background-color:/*#eaffea*/#D6FFD6;">📌</button></a>`)
var topic_39 []byte = []byte(`<a href='/topic/stick/submit/`)
var topic_40 []byte = []byte(`' class="mod_button" style="font-weight:normal;" title="Pin Topic"><button class="username">📌</button></a>&nbsp;`)
var topic_41 []byte = []byte(`
		<a href="/report/submit/`)
var topic_42 []byte = []byte(`?session=`)
var topic_43 []byte = []byte(`&type=topic" class="mod_button report_item" style="font-weight:normal;" title="Flag Topic"><button class="username">🚩</button></a>&nbsp;
		`)
var topic_44 []byte = []byte(`<a class="username hide_on_micro" style="float: right;color:#505050;border-left:none;padding-left:5px;padding-right:5px;font-size:17px;">`)
var topic_45 []byte = []byte(`</a><a class="username hide_on_micro" style="color:#505050;float:right;opacity:0.85;margin-left:5px;" title="Like Count">😀</a>`)
var topic_46 []byte = []byte(`<a class="username hide_on_micro" style="float:right;color:#505050;font-size:16px;">`)
var topic_47 []byte = []byte(`</a>`)
var topic_48 []byte = []byte(`<a class="username hide_on_micro" style="float: right;color:#505050;border-left:none;padding-left:5px;padding-right:5px;font-size:17px;">`)
var topic_49 []byte = []byte(`</a><a class="username hide_on_micro" style="color:#505050;float:right;opacity:0.85;" title="Level">👑</a>`)
var topic_50 []byte = []byte(`
	</div>
</div><br />
<div class="rowblock post_container" style="overflow: hidden;">`)
var topic_51 []byte = []byte(`
	<div class="rowitem rowhead passive deletable_block editable_parent post_item" style="`)
var topic_52 []byte = []byte(`background-image:url(`)
var topic_53 []byte = []byte(`), url(/static/white-dot.jpg);background-position: 0px `)
var topic_54 []byte = []byte(`-1`)
var topic_55 []byte = []byte(`0px;background-repeat:no-repeat, repeat-y;background-size:128px;padding-left:136px;`)
var topic_56 []byte = []byte(`">
		<p class="editable_block user_content" style="margin: 0;padding: 0;">`)
var topic_57 []byte = []byte(`</p><br /><br />
		<a href="/user/`)
var topic_58 []byte = []byte(`" class="username real_username">`)
var topic_59 []byte = []byte(`</a>&nbsp;
		`)
var topic_60 []byte = []byte(`<a href="/reply/like/submit/`)
var topic_61 []byte = []byte(`" class="mod_button" title="Love it" style="color:#202020;"><button class="username" style="`)
var topic_62 []byte = []byte(`background-color:/*#eaffea*/#D6FFD6;`)
var topic_63 []byte = []byte(`">😀</button></a>&nbsp;`)
var topic_64 []byte = []byte(`<a href="/reply/edit/submit/`)
var topic_65 []byte = []byte(`" class="mod_button" title="Edit Reply"><button class="username edit_item">🖊️</button></a>&nbsp;`)
var topic_66 []byte = []byte(`<a href="/reply/delete/submit/`)
var topic_67 []byte = []byte(`" class="mod_button" title="Delete Reply"><button class="username delete_item">🗑️</button></a>&nbsp;`)
var topic_68 []byte = []byte(`
		<a href="/report/submit/`)
var topic_69 []byte = []byte(`?session=`)
var topic_70 []byte = []byte(`&type=reply" class="mod_button" title="Flag Reply"><button class="username report_item">🚩</button></a>&nbsp;
		`)
var topic_71 []byte = []byte(`<a class="username hide_on_micro" style="float: right;color:#505050;border-left:none;padding-left:5px;padding-right:5px;font-size:17px;">`)
var topic_72 []byte = []byte(`</a><a class="username hide_on_micro" style="color:#505050;float:right;opacity:0.85;margin-left:5px;" title="Like Count">😀</a>`)
var topic_73 []byte = []byte(`<a class="username hide_on_micro" style="float: right;color:#505050;font-size:16px;">`)
var topic_74 []byte = []byte(`</a>`)
var topic_75 []byte = []byte(`<a class="username hide_on_micro" style="float: right;color:#505050;border-left:none;padding-left:5px;padding-right:5px;font-size:17px;">`)
var topic_76 []byte = []byte(`</a><a class="username hide_on_micro" style="color:#505050;float:right;opacity:0.85;" title="Level">👑`)
var topic_77 []byte = []byte(`</a>
	</div>
`)
var topic_78 []byte = []byte(`</div>
`)
var topic_79 []byte = []byte(`
<div class="rowblock">
	<form action="/reply/create/" method="post">
		<input name="tid" value='`)
var topic_80 []byte = []byte(`' type="hidden" />
		<div class="formrow">
			<div class="formitem"><textarea name="reply-content" placeholder="Insert reply here"></textarea></div>
		</div>
		<div class="formrow">
			<div class="formitem"><button name="reply-button" class="formbutton">Create Reply</button></div>
		</div>
	</form>
</div>
`)
var footer_0 []byte = []byte(`			<!--<link rel="stylesheet" href="https://use.fontawesome.com/8670aa03ca.css">-->
		</div><div style="clear: both;"></div></div></div>
	</body>
</html>`)
var topic_alt_0 []byte = []byte(`<div id="prevFloat" class="prev_button"><a class="prev_link" href="/topic/`)
var topic_alt_1 []byte = []byte(`?page=`)
var topic_alt_2 []byte = []byte(`">&lt;</a></div>`)
var topic_alt_3 []byte = []byte(`<link rel="prerender" href="/topic/`)
var topic_alt_4 []byte = []byte(`?page=`)
var topic_alt_5 []byte = []byte(`" />
<div id="nextFloat" class="next_button"><a class="next_link" href="/topic/`)
var topic_alt_6 []byte = []byte(`?page=`)
var topic_alt_7 []byte = []byte(`">&gt;</a></div>`)
var topic_alt_8 []byte = []byte(`
<div class="rowblock topic_block">
	<form action='/topic/edit/submit/`)
var topic_alt_9 []byte = []byte(`' method="post">
		<div class="rowitem topic_item rowhead`)
var topic_alt_10 []byte = []byte(` topic_sticky_head`)
var topic_alt_11 []byte = []byte(` topic_closed_head`)
var topic_alt_12 []byte = []byte(`">
			<a class='topic_name hide_on_edit'>`)
var topic_alt_13 []byte = []byte(`</a> 
			`)
var topic_alt_14 []byte = []byte(`<span class='username hide_on_micro topic_status_e topic_status_closed hide_on_edit' title='Status: Closed' style="font-weight:normal;float: right;position:relative;top:-5px;">&#x1F512;&#xFE0E</span>`)
var topic_alt_15 []byte = []byte(`
			<input class='show_on_edit topic_name_input' name="topic_name" value='`)
var topic_alt_16 []byte = []byte(`' type="text" />
			`)
var topic_alt_17 []byte = []byte(`<select name="topic_status" class='show_on_edit topic_status_input' style='float: right;'>
				<option>open</option>
				<option>closed</option>
			</select>`)
var topic_alt_18 []byte = []byte(`
			<button name="topic-button" class="formbutton show_on_edit submit_edit">Update</button>
			`)
var topic_alt_19 []byte = []byte(`
		</div>
	</form>
</div>
<style type="text/css">.rowitem:last-child .content_container { margin-bottom: 5px !important; }</style>
<div class="rowblock post_container" style="border-top: none;">
	<div class="rowitem passive deletable_block editable_parent post_item" style="background-color: #eaeaea;padding-top: 4px;padding-left: 5px;clear: both;border-bottom: none;padding-right: 4px;padding-bottom: 2px;">
		<div class="userinfo">
			<div class="avatar_item" style="background-image: url(`)
var topic_alt_20 []byte = []byte(`), url(/static/white-dot.jpg);background-position: 0px -10px;">&nbsp;</div>
			<a href="/user/`)
var topic_alt_21 []byte = []byte(`" class="the_name">`)
var topic_alt_22 []byte = []byte(`</a>
			`)
var topic_alt_23 []byte = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag">`)
var topic_alt_24 []byte = []byte(`</div><div class="tag_post"></div></div>`)
var topic_alt_25 []byte = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag post_level">Level `)
var topic_alt_26 []byte = []byte(`</div><div class="tag_post"></div></div>`)
var topic_alt_27 []byte = []byte(`
		</div>
		<div class="content_container">
			<div class="hide_on_edit topic_content user_content">`)
var topic_alt_28 []byte = []byte(`</div>
			<textarea name="topic_content" class="show_on_edit topic_content_input">`)
var topic_alt_29 []byte = []byte(`</textarea>
			<div class="button_container">
				`)
var topic_alt_30 []byte = []byte(`<a href="/topic/like/submit/`)
var topic_alt_31 []byte = []byte(`" class="action_button">+1</a>`)
var topic_alt_32 []byte = []byte(`<a href="/topic/edit/`)
var topic_alt_33 []byte = []byte(`" class="action_button open_edit">Edit</a>`)
var topic_alt_34 []byte = []byte(`<a href="/topic/delete/submit/`)
var topic_alt_35 []byte = []byte(`" class="action_button delete_item">Delete</a>`)
var topic_alt_36 []byte = []byte(`<a href='/topic/unstick/submit/`)
var topic_alt_37 []byte = []byte(`' class="action_button">Unpin</a>`)
var topic_alt_38 []byte = []byte(`<a href='/topic/stick/submit/`)
var topic_alt_39 []byte = []byte(`' class="action_button">Pin</a>`)
var topic_alt_40 []byte = []byte(`
				<a href="/report/submit/`)
var topic_alt_41 []byte = []byte(`?session=`)
var topic_alt_42 []byte = []byte(`&type=topic" class="action_button report_item">Report</a>
				`)
var topic_alt_43 []byte = []byte(`<a href="#" title="IP Address" class="action_button action_button_right ip_item hide_on_mobile">`)
var topic_alt_44 []byte = []byte(`</a>`)
var topic_alt_45 []byte = []byte(`<a class="action_button action_button_right hide_on_micro">`)
var topic_alt_46 []byte = []byte(` up</a>`)
var topic_alt_47 []byte = []byte(`
			</div>
		</div><div style="clear:both;"></div>
	</div>
	`)
var topic_alt_48 []byte = []byte(`
	<div class="rowitem passive deletable_block editable_parent post_item">
		<div class="userinfo">
			<div class="avatar_item" style="background-image: url(`)
var topic_alt_49 []byte = []byte(`), url(/static/white-dot.jpg);background-position: 0px -10px;">&nbsp;</div>
			<a href="/user/`)
var topic_alt_50 []byte = []byte(`" class="the_name">`)
var topic_alt_51 []byte = []byte(`</a>
			`)
var topic_alt_52 []byte = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag">`)
var topic_alt_53 []byte = []byte(`</div><div class="tag_post"></div></div>`)
var topic_alt_54 []byte = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag post_level">Level `)
var topic_alt_55 []byte = []byte(`</div><div class="tag_post"></div></div>`)
var topic_alt_56 []byte = []byte(`
		</div>
		<div class="content_container">
			<div class="editable_block user_content">`)
var topic_alt_57 []byte = []byte(`</div>
			<div class="button_container">
				`)
var topic_alt_58 []byte = []byte(`<a href="/reply/like/submit/`)
var topic_alt_59 []byte = []byte(`" class="action_button">+1</a>`)
var topic_alt_60 []byte = []byte(`<a href="/reply/edit/submit/`)
var topic_alt_61 []byte = []byte(`" class="action_button edit_item">Edit</a>`)
var topic_alt_62 []byte = []byte(`<a href="/reply/delete/submit/`)
var topic_alt_63 []byte = []byte(`" class="action_button delete_item">Delete</a>`)
var topic_alt_64 []byte = []byte(`
				<a href="/report/submit/`)
var topic_alt_65 []byte = []byte(`?session=`)
var topic_alt_66 []byte = []byte(`&type=reply" class="action_button report_item">Report</a>
				`)
var topic_alt_67 []byte = []byte(`<a href="#" title="IP Address" class="action_button action_button_right ip_item hide_on_mobile">`)
var topic_alt_68 []byte = []byte(`</a>`)
var topic_alt_69 []byte = []byte(`<a class="action_button action_button_right hide_on_micro">`)
var topic_alt_70 []byte = []byte(` up</a>`)
var topic_alt_71 []byte = []byte(`
			</div>
		</div>
		<div style="clear:both;"></div>
	</div>
`)
var topic_alt_72 []byte = []byte(`</div>
`)
var topic_alt_73 []byte = []byte(`
<div class="rowblock" style="border-top: none;">
	<form action="/reply/create/" method="post">
		<input name="tid" value='`)
var topic_alt_74 []byte = []byte(`' type="hidden" />
		<div class="formrow">
			<div class="formitem"><textarea name="reply-content" placeholder="Insert reply here"></textarea></div>
		</div>
		<div class="formrow">
			<div class="formitem"><button name="reply-button" class="formbutton">Create Reply</button></div>
		</div>
	</form>
</div>
`)
var profile_0 []byte = []byte(`
<div class="colblock_left" style="max-width: 220px;">
	<div class="rowitem" style="padding: 0;"><img src="`)
var profile_1 []byte = []byte(`" style="max-width: 100%;margin: 0;display: block;" /></div>
	<div class="rowitem" style="text-transform: capitalize;">
	<span style="font-size: 18px;">`)
var profile_2 []byte = []byte(`</span>`)
var profile_3 []byte = []byte(`<span class="username" style="float: right;font-weight: normal;">`)
var profile_4 []byte = []byte(`</span>`)
var profile_5 []byte = []byte(`
	</div>
	<div class="rowitem passive">
		<a class="username">Add Friend</a>
		`)
var profile_6 []byte = []byte(`<a href="/users/unban/`)
var profile_7 []byte = []byte(`?session=`)
var profile_8 []byte = []byte(`" class="username">Unban</a>`)
var profile_9 []byte = []byte(`<a href="/users/ban/`)
var profile_10 []byte = []byte(`?session=`)
var profile_11 []byte = []byte(`" class="username">Ban</a>`)
var profile_12 []byte = []byte(`
		<a href="/report/submit/`)
var profile_13 []byte = []byte(`?session=`)
var profile_14 []byte = []byte(`&type=user" class="username report_item">Report</a>
	</div>
</div>
<div class="colblock_right">
	<div class="rowitem rowhead"><a>Comments</a></div>
</div>
<div class="colblock_right" style="overflow: hidden;border-top: none;">`)
var profile_15 []byte = []byte(`
<div class="rowitem passive deletable_block editable_parent simple" style="`)
var profile_16 []byte = []byte(`background-image: url(`)
var profile_17 []byte = []byte(`), url(/static/white-dot.jpg);background-position: 0px `)
var profile_18 []byte = []byte(`-1`)
var profile_19 []byte = []byte(`0px;background-repeat: no-repeat, repeat-y;background-size: 128px;padding-left: 136px;`)
var profile_20 []byte = []byte(`">
		<span class="editable_block user_content simple">`)
var profile_21 []byte = []byte(`</span>
		<br /><br />
		<a href="/user/`)
var profile_22 []byte = []byte(`" class="username">`)
var profile_23 []byte = []byte(`</a>
		`)
var profile_24 []byte = []byte(`<a href="/profile/reply/edit/submit/`)
var profile_25 []byte = []byte(`"><button class="username edit_item">Edit</button></a>
		<a href="/profile/reply/delete/submit/`)
var profile_26 []byte = []byte(`"><button class="username delete_item">Delete</button></a>`)
var profile_27 []byte = []byte(`
		<a href="/report/submit/`)
var profile_28 []byte = []byte(`?session=`)
var profile_29 []byte = []byte(`&type=user-reply"><button class="username report_item">Report</button></a>
		`)
var profile_30 []byte = []byte(`<a class="username hide_on_mobile" style="float: right;">`)
var profile_31 []byte = []byte(`</a>`)
var profile_32 []byte = []byte(`
	</div>
`)
var profile_33 []byte = []byte(`</div>
<div class="colblock_right" style="border-top: none;">
`)
var profile_34 []byte = []byte(`
<form action="/profile/reply/create/" method="post">
	<input name="uid" value='`)
var profile_35 []byte = []byte(`' type="hidden" />
	<div class="formrow">
		<div class="formitem"><textarea name="reply-content" placeholder="Insert reply here"></textarea></div>
	</div>
	<div class="formrow">
		<div class="formitem"><button name="reply-button" class="formbutton">Create Reply</button></div>
	</div>
</form>
`)
var profile_36 []byte = []byte(`
</div>
`)
var forums_0 []byte = []byte(`
<div class="rowblock">
	`)
var forums_1 []byte = []byte(`<div class="rowitem">
		<a href="/forum/`)
var forums_2 []byte = []byte(`" style="font-size: 20px;position:relative;top: -2px;font-weight: normal;text-transform: none;">`)
var forums_3 []byte = []byte(`</a>
		<a href="/topic/`)
var forums_4 []byte = []byte(`" style="font-weight: normal;text-transform: none;float: right;">`)
var forums_5 []byte = []byte(` <small style="font-size: 12px;">`)
var forums_6 []byte = []byte(`</small></a>
	</div>
	`)
var forums_7 []byte = []byte(`<div class="rowitem passive">You don't have access to any forums.</div>`)
var forums_8 []byte = []byte(`
</div>
`)
var topics_0 []byte = []byte(`
<div class="rowblock">
	<div class="rowitem rowhead"><a>Topic List</a></div>
</div>
<div class="rowblock">
	`)
var topics_1 []byte = []byte(`<div class="rowitem passive" style="`)
var topics_2 []byte = []byte(`background-image: url(`)
var topics_3 []byte = []byte(`);background-position: left;background-repeat: no-repeat;background-size: 64px;padding-left: 72px;`)
var topics_4 []byte = []byte(`background-color: #FFFFCC;`)
var topics_5 []byte = []byte(`background-color: #eaeaea;`)
var topics_6 []byte = []byte(`">
		<a href="/topic/`)
var topics_7 []byte = []byte(`">`)
var topics_8 []byte = []byte(`</a> `)
var topics_9 []byte = []byte(`<a href="/forum/`)
var topics_10 []byte = []byte(`" style="font-size:12px;">`)
var topics_11 []byte = []byte(`</a> `)
var topics_12 []byte = []byte(`<span class="username topic_status_e topic_status_closed" style="float: right;position:relative;top:-5px;margin-left:8px;" title="Status: Closed">&#x1F512;&#xFE0E</span>`)
var topics_13 []byte = []byte(`
		<a style="float: right;font-size:12px;">`)
var topics_14 []byte = []byte(`</a>
	</div>
	`)
var topics_15 []byte = []byte(`<div class="rowitem passive">There aren't any topics yet.`)
var topics_16 []byte = []byte(` <a href="/topics/create/">Start one?</a>`)
var topics_17 []byte = []byte(`</div>`)
var topics_18 []byte = []byte(`
</div>
`)
var forum_0 []byte = []byte(`<div id="prevFloat" class="prev_button"><a class="prev_link" href="/forum/`)
var forum_1 []byte = []byte(`?page=`)
var forum_2 []byte = []byte(`">&lt;</a></div>`)
var forum_3 []byte = []byte(`<link rel="prerender" href="/forum/`)
var forum_4 []byte = []byte(`?page=`)
var forum_5 []byte = []byte(`" />
<div id="nextFloat" class="next_button"><a class="next_link" href="/forum/`)
var forum_6 []byte = []byte(`?page=`)
var forum_7 []byte = []byte(`">&gt;</a></div>`)
var forum_8 []byte = []byte(`
<div class="rowblock">
	<div class="rowitem rowhead"><a>`)
var forum_9 []byte = []byte(`</a>
    `)
var forum_10 []byte = []byte(`<span class='username' title='No Permissions' style="font-weight:normal;float: right;position:relative;top:-5px;">&#x1F512;&#xFE0E</span>`)
var forum_11 []byte = []byte(`<a href="/topics/create/`)
var forum_12 []byte = []byte(`" class='username' style="float: right;position:relative;top:-5px;">New Topic</a>`)
var forum_13 []byte = []byte(`</div>
</div>
<div class="rowblock">
	`)
var forum_14 []byte = []byte(`<div class="rowitem passive" style="`)
var forum_15 []byte = []byte(`background-image: url(`)
var forum_16 []byte = []byte(`);background-position: left;background-repeat: no-repeat;background-size: 64px;padding-left: 72px;`)
var forum_17 []byte = []byte(`background-color: #FFFFCC;`)
var forum_18 []byte = []byte(`background-color: #eaeaea;`)
var forum_19 []byte = []byte(`">
		<a href="/topic/`)
var forum_20 []byte = []byte(`">`)
var forum_21 []byte = []byte(`</a> `)
var forum_22 []byte = []byte(`<span class="username topic_status_e topic_status_closed" title="Status: Closed" style="float: right;position:relative;top:-5px;margin-left:8px;">&#x1F512;&#xFE0E</span>`)
var forum_23 []byte = []byte(`
		<a style="float: right;font-size:12px;">`)
var forum_24 []byte = []byte(`</a>
	</div>
	`)
var forum_25 []byte = []byte(`<div class="rowitem passive">There aren't any topics in this forum yet.`)
var forum_26 []byte = []byte(` <a href="/topics/create/`)
var forum_27 []byte = []byte(`">Start one?</a>`)
var forum_28 []byte = []byte(`</div>`)
var forum_29 []byte = []byte(`
</div>
`)
