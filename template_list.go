package main

var header_0 []byte = []byte(`<!doctype html>
<html lang="en">
	<head>
		<title>`)
var header_1 []byte = []byte(`</title>
		<link href="/static/main.css" rel="stylesheet" type="text/css">
		`)
var header_2 []byte = []byte(`
		<link href="/static/`)
var header_3 []byte = []byte(`" rel="stylesheet" type="text/css">
		`)
var header_4 []byte = []byte(`
		<script type="text/javascript" src="/static/jquery-3.1.1.min.js"></script>
		`)
var header_5 []byte = []byte(`
		<script type="text/javascript" src="/static/`)
var header_6 []byte = []byte(`"></script>
		`)
var header_7 []byte = []byte(`
		<script type="text/javascript">var session = "`)
var header_8 []byte = []byte(`";
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
		<li class="menu_left menu_logout"><a href="/accounts/logout/?session=`)
var menu_5 []byte = []byte(`">Logout</a></li>
		`)
var menu_6 []byte = []byte(`
		<li class="menu_left menu_register"><a href="/accounts/create/">Register</a></li>
		<li class="menu_left menu_login"><a href="/accounts/login/">Login</a></li>
		`)
var menu_7 []byte = []byte(`
		<li id="general_alerts" class="menu_right menu_alerts">
			<div class="alert_bell">🔔︎</div>
			<div class="alert_counter"></div>
			<div class="alertList"></div>
		</li>
	</ul>
	</div>
	</div>
	<div style="clear: both;"></div>
</div>
`)
var header_9 []byte = []byte(`
<div id="back"><div id="main" `)
var header_10 []byte = []byte(`class="shrink_main"`)
var header_11 []byte = []byte(`>
`)
var header_12 []byte = []byte(`<div class="alert">`)
var header_13 []byte = []byte(`</div>`)
var topic_0 []byte = []byte(`<div id="prevFloat" class="prev_button"><a class="prev_link" href="/topic/`)
var topic_1 []byte = []byte(`?page=`)
var topic_2 []byte = []byte(`">&lt;</a></div>`)
var topic_3 []byte = []byte(`<link rel="prerender" href="/topic/`)
var topic_4 []byte = []byte(`?page=`)
var topic_5 []byte = []byte(`" />
<div id="nextFloat" class="next_button">
	<a class="next_link" href="/topic/`)
var topic_6 []byte = []byte(`?page=`)
var topic_7 []byte = []byte(`">&gt;</a>
</div>`)
var topic_8 []byte = []byte(`

<div class="rowblock topic_block">
	<form action='/topic/edit/submit/`)
var topic_9 []byte = []byte(`' method="post">
		<div class="rowitem rowhead topic_item"`)
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
<div class="rowblock post_container top_post">
	<div class="rowitem passive editable_parent post_item" style="border-bottom: none;`)
var topic_20 []byte = []byte(`background-image:url(`)
var topic_21 []byte = []byte(`), url(/static/white-dot.jpg);background-position: 0px `)
var topic_22 []byte = []byte(`-1`)
var topic_23 []byte = []byte(`0px;background-repeat:no-repeat, repeat-y;background-size:128px;padding-left:136px;`)
var topic_24 []byte = []byte(`">
		<p class="hide_on_edit topic_content user_content" style="margin:0;padding:0;">`)
var topic_25 []byte = []byte(`</p>
		<textarea name="topic_content" class="show_on_edit topic_content_input">`)
var topic_26 []byte = []byte(`</textarea>
		
		<span class="controls">
		
		<a href="/user/`)
var topic_27 []byte = []byte(`" class="username real_username">`)
var topic_28 []byte = []byte(`</a>&nbsp;&nbsp;
		
		`)
var topic_29 []byte = []byte(`<a href="/topic/like/submit/`)
var topic_30 []byte = []byte(`" class="mod_button" title="Love it" style="color:#202020;">
		<button class="username like_label" style="`)
var topic_31 []byte = []byte(`background-color:/*#eaffea*/#D6FFD6;`)
var topic_32 []byte = []byte(`"></button></a>`)
var topic_33 []byte = []byte(`<a href='/topic/edit/`)
var topic_34 []byte = []byte(`' class="mod_button open_edit" style="font-weight:normal;" title="Edit Topic"><button class="username edit_label"></button></a>`)
var topic_35 []byte = []byte(`<a href='/topic/delete/submit/`)
var topic_36 []byte = []byte(`' class="mod_button" style="font-weight:normal;" title="Delete Topic"><button class="username trash_label"></button></a>`)
var topic_37 []byte = []byte(`<a class="mod_button" href='/topic/unstick/submit/`)
var topic_38 []byte = []byte(`' style="font-weight:normal;" title="Unpin Topic"><button class="username unpin_label"></button></a>`)
var topic_39 []byte = []byte(`<a href='/topic/stick/submit/`)
var topic_40 []byte = []byte(`' class="mod_button" style="font-weight:normal;" title="Pin Topic"><button class="username pin_label"></button></a>`)
var topic_41 []byte = []byte(`
		
		<a class="mod_button" href="/report/submit/`)
var topic_42 []byte = []byte(`?session=`)
var topic_43 []byte = []byte(`&type=topic" class="mod_button report_item" style="font-weight:normal;" title="Flag Topic"><button class="username flag_label"></button></a>
		
		`)
var topic_44 []byte = []byte(`<a class="username hide_on_micro like_count">`)
var topic_45 []byte = []byte(`</a><a class="username hide_on_micro like_count_label" title="Like Count"></a>`)
var topic_46 []byte = []byte(`<a class="username hide_on_micro" style="float:right;color:#505050;font-size:16px;">`)
var topic_47 []byte = []byte(`</a>`)
var topic_48 []byte = []byte(`<a class="username hide_on_micro level">`)
var topic_49 []byte = []byte(`</a><a class="username hide_on_micro level_label" style="color:#505050;float:right;opacity:0.85;" title="Level"></a>`)
var topic_50 []byte = []byte(`
		
		</span>
	</div>
</div>
<div class="rowblock post_container" style="overflow: hidden;">`)
var topic_51 []byte = []byte(`
	<div class="rowitem passive deletable_block editable_parent post_item action_item">
		<span class="action_icon" style="font-size: 18px;padding-right: 5px;">`)
var topic_52 []byte = []byte(`</span>
		<span>`)
var topic_53 []byte = []byte(`</span>
	</div>
`)
var topic_54 []byte = []byte(`
	<div class="rowitem passive deletable_block editable_parent post_item" style="`)
var topic_55 []byte = []byte(`background-image:url(`)
var topic_56 []byte = []byte(`), url(/static/white-dot.jpg);background-position: 0px `)
var topic_57 []byte = []byte(`-1`)
var topic_58 []byte = []byte(`0px;background-repeat:no-repeat, repeat-y;background-size:128px;padding-left:136px;`)
var topic_59 []byte = []byte(`">
		<p class="editable_block user_content" style="margin:0;padding:0;">`)
var topic_60 []byte = []byte(`</p>
		
		<span class="controls">
		
		<a href="/user/`)
var topic_61 []byte = []byte(`" class="username real_username">`)
var topic_62 []byte = []byte(`</a>&nbsp;&nbsp;
		
		`)
var topic_63 []byte = []byte(`<a href="/reply/like/submit/`)
var topic_64 []byte = []byte(`" class="mod_button" title="Love it" style="color:#202020;"><button class="username like_label" style="`)
var topic_65 []byte = []byte(`background-color:/*#eaffea*/#D6FFD6;`)
var topic_66 []byte = []byte(`"></button></a>`)
var topic_67 []byte = []byte(`<a href="/reply/edit/submit/`)
var topic_68 []byte = []byte(`" class="mod_button" title="Edit Reply"><button class="username edit_item edit_label"></button></a>`)
var topic_69 []byte = []byte(`<a href="/reply/delete/submit/`)
var topic_70 []byte = []byte(`" class="mod_button" title="Delete Reply"><button class="username delete_item trash_label"></button></a>`)
var topic_71 []byte = []byte(`
		
		<a class="mod_button" href="/report/submit/`)
var topic_72 []byte = []byte(`?session=`)
var topic_73 []byte = []byte(`&type=reply" class="mod_button report_item" title="Flag Reply"><button class="username report_item flag_label"></button></a>
		
		`)
var topic_74 []byte = []byte(`<a class="username hide_on_micro like_count">`)
var topic_75 []byte = []byte(`</a><a class="username hide_on_micro like_count_label" title="Like Count"></a>`)
var topic_76 []byte = []byte(`<a class="username hide_on_micro" style="float: right;color:#505050;font-size:16px;">`)
var topic_77 []byte = []byte(`</a>`)
var topic_78 []byte = []byte(`<a class="username hide_on_micro level">`)
var topic_79 []byte = []byte(`</a><a class="username hide_on_micro level_label" style="color:#505050;float:right;opacity:0.85;" title="Level">`)
var topic_80 []byte = []byte(`</a>
		
		</span>
	</div>
`)
var topic_81 []byte = []byte(`</div>

`)
var topic_82 []byte = []byte(`
<div class="rowblock">
	<form action="/reply/create/" method="post">
		<input name="tid" value='`)
var topic_83 []byte = []byte(`' type="hidden" />
		<div class="formrow">
			<div class="formitem"><textarea name="reply-content" placeholder="Insert reply here"></textarea></div>
		</div>
		<div class="formrow">
			<div class="formitem"><button name="reply-button" class="formbutton">Create Reply</button></div>
		</div>
	</form>
</div>
`)
var footer_0 []byte = []byte(`					</div>
				`)
var footer_1 []byte = []byte(`<div class="sidebar">`)
var footer_2 []byte = []byte(`</div>`)
var footer_3 []byte = []byte(`
				<div style="clear: both;"></div>
			</div>
		</div>
	</body>
</html>
`)
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
var topic_alt_45 []byte = []byte(`
				<a class="action_button action_button_right hide_on_mobile">`)
var topic_alt_46 []byte = []byte(`</a>
				`)
var topic_alt_47 []byte = []byte(`<a class="action_button action_button_right hide_on_micro">`)
var topic_alt_48 []byte = []byte(` up</a>`)
var topic_alt_49 []byte = []byte(`
			</div>
		</div><div style="clear:both;"></div>
	</div>
	`)
var topic_alt_50 []byte = []byte(`
	<div class="rowitem passive deletable_block editable_parent post_item `)
var topic_alt_51 []byte = []byte(`action_item`)
var topic_alt_52 []byte = []byte(`">
		<div class="userinfo">
			<div class="avatar_item" style="background-image: url(`)
var topic_alt_53 []byte = []byte(`), url(/static/white-dot.jpg);background-position: 0px -10px;">&nbsp;</div>
			<a href="/user/`)
var topic_alt_54 []byte = []byte(`" class="the_name">`)
var topic_alt_55 []byte = []byte(`</a>
			`)
var topic_alt_56 []byte = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag">`)
var topic_alt_57 []byte = []byte(`</div><div class="tag_post"></div></div>`)
var topic_alt_58 []byte = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag post_level">Level `)
var topic_alt_59 []byte = []byte(`</div><div class="tag_post"></div></div>`)
var topic_alt_60 []byte = []byte(`
		</div>
		<div class="content_container" `)
var topic_alt_61 []byte = []byte(`style="margin-left: 0px;"`)
var topic_alt_62 []byte = []byte(`>
			`)
var topic_alt_63 []byte = []byte(`
				<span class="action_icon" style="font-size: 18px;padding-right: 5px;">`)
var topic_alt_64 []byte = []byte(`</span>
				<span>`)
var topic_alt_65 []byte = []byte(`</span>
			`)
var topic_alt_66 []byte = []byte(`
			<div class="editable_block user_content">`)
var topic_alt_67 []byte = []byte(`</div>
			<div class="button_container">
				`)
var topic_alt_68 []byte = []byte(`<a href="/reply/like/submit/`)
var topic_alt_69 []byte = []byte(`" class="action_button">+1</a>`)
var topic_alt_70 []byte = []byte(`<a href="/reply/edit/submit/`)
var topic_alt_71 []byte = []byte(`" class="action_button edit_item">Edit</a>`)
var topic_alt_72 []byte = []byte(`<a href="/reply/delete/submit/`)
var topic_alt_73 []byte = []byte(`" class="action_button delete_item">Delete</a>`)
var topic_alt_74 []byte = []byte(`
					<a href="/report/submit/`)
var topic_alt_75 []byte = []byte(`?session=`)
var topic_alt_76 []byte = []byte(`&type=reply" class="action_button report_item">Report</a>
					`)
var topic_alt_77 []byte = []byte(`<a href="#" title="IP Address" class="action_button action_button_right ip_item hide_on_mobile">`)
var topic_alt_78 []byte = []byte(`</a>`)
var topic_alt_79 []byte = []byte(`
				<a class="action_button action_button_right hide_on_mobile">`)
var topic_alt_80 []byte = []byte(`</a>
				`)
var topic_alt_81 []byte = []byte(`<a class="action_button action_button_right hide_on_micro">`)
var topic_alt_82 []byte = []byte(` up</a>`)
var topic_alt_83 []byte = []byte(`
			</div>
			`)
var topic_alt_84 []byte = []byte(`
		</div>
		<div style="clear:both;"></div>
	</div>
`)
var topic_alt_85 []byte = []byte(`</div>
`)
var topic_alt_86 []byte = []byte(`
<div class="rowblock" style="border-top: none;">
	<form action="/reply/create/" method="post">
		<input name="tid" value='`)
var topic_alt_87 []byte = []byte(`' type="hidden" />
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

<div id="profile_left_pane" class="colblock_left" style="max-width: 220px;">
	<div class="rowitem" style="padding: 0;">
		<img src="`)
var profile_1 []byte = []byte(`" style="max-width: 100%;margin: 0;display: block;" />
	</div>
	<div class="rowitem">
	<span style="font-size: 18px;">`)
var profile_2 []byte = []byte(`</span>`)
var profile_3 []byte = []byte(`<span class="username" style="float: right;font-weight: normal;">`)
var profile_4 []byte = []byte(`</span>`)
var profile_5 []byte = []byte(`
	</div>
	<div class="rowitem passive">
		<a class="profile_menu_item">Add Friend</a>
	</div>
	`)
var profile_6 []byte = []byte(`<div class="rowitem passive">
		`)
var profile_7 []byte = []byte(`<a href="/users/unban/`)
var profile_8 []byte = []byte(`?session=`)
var profile_9 []byte = []byte(`" class="profile_menu_item">Unban</a>
		`)
var profile_10 []byte = []byte(`<a href="/users/ban/`)
var profile_11 []byte = []byte(`?session=`)
var profile_12 []byte = []byte(`" class="profile_menu_item">Ban</a>`)
var profile_13 []byte = []byte(`
	</div>`)
var profile_14 []byte = []byte(`
	<div class="rowitem passive">
		<a href="/report/submit/`)
var profile_15 []byte = []byte(`?session=`)
var profile_16 []byte = []byte(`&type=user" class="profile_menu_item report_item">Report</a>
	</div>
</div>

<div class="colblock_right" style="width: calc(95% - 210px);">
	<div class="rowitem rowhead"><a>Comments</a></div>
</div>
<div id="profile_comments" class="colblock_right" style="overflow: hidden;border-top: none;width:calc(95% - 210px);">`)
var profile_17 []byte = []byte(`
<div class="rowitem passive deletable_block editable_parent simple" style="`)
var profile_18 []byte = []byte(`background-image: url(`)
var profile_19 []byte = []byte(`), url(/static/white-dot.jpg);background-position: 0px `)
var profile_20 []byte = []byte(`-1`)
var profile_21 []byte = []byte(`0px;background-repeat: no-repeat, repeat-y;background-size: 128px;padding-left: 136px;`)
var profile_22 []byte = []byte(`">
		<span class="editable_block user_content simple">`)
var profile_23 []byte = []byte(`</span><br /><br />
		<a href="/user/`)
var profile_24 []byte = []byte(`" class="real_username username">`)
var profile_25 []byte = []byte(`</a>&nbsp;&nbsp;

		`)
var profile_26 []byte = []byte(`<a href="/profile/reply/edit/submit/`)
var profile_27 []byte = []byte(`" class="mod_button" title="Edit Item"><button class="username edit_item edit_label"></button></a>

		<a href="/profile/reply/delete/submit/`)
var profile_28 []byte = []byte(`" class="mod_button" title="Delete Item"><button class="username delete_item trash_label"></button></a>`)
var profile_29 []byte = []byte(`

		<a class="mod_button" href="/report/submit/`)
var profile_30 []byte = []byte(`?session=`)
var profile_31 []byte = []byte(`&type=user-reply"><button class="username report_item flag_label"></button></a>

		`)
var profile_32 []byte = []byte(`<a class="username hide_on_mobile" style="float: right;">`)
var profile_33 []byte = []byte(`</a>`)
var profile_34 []byte = []byte(`
	</div>
`)
var profile_35 []byte = []byte(`</div>

<div class="colblock_right" style="border-top: none;width: calc(95% - 210px);">
`)
var profile_36 []byte = []byte(`
<form action="/profile/reply/create/" method="post">
	<input name="uid" value='`)
var profile_37 []byte = []byte(`' type="hidden" />
	<div class="formrow">
		<div class="formitem"><textarea name="reply-content" placeholder="Insert reply here"></textarea></div>
	</div>
	<div class="formrow">
		<div class="formitem"><button name="reply-button" class="formbutton">Create Reply</button></div>
	</div>
</form>
`)
var profile_38 []byte = []byte(`
</div>

`)
var forums_0 []byte = []byte(`
<div class="rowblock opthead">
	<div class="rowitem rowhead"><a>Forums</a></div>
</div>
<div class="rowblock">
	`)
var forums_1 []byte = []byte(`<div class="rowitem `)
var forums_2 []byte = []byte(`datarow`)
var forums_3 []byte = []byte(`">
		`)
var forums_4 []byte = []byte(`<span style="float: left;">
			<a href="/forum/`)
var forums_5 []byte = []byte(`" style="">`)
var forums_6 []byte = []byte(`</a>
			<br /><span class="rowsmall">`)
var forums_7 []byte = []byte(`</span>
		</span>`)
var forums_8 []byte = []byte(`<span style="float: left;padding-top: 8px;font-size: 18px;">
			<a href="/forum/`)
var forums_9 []byte = []byte(`">`)
var forums_10 []byte = []byte(`</a>
		</span>`)
var forums_11 []byte = []byte(`<span style="float: left;">
			<a href="/forum/`)
var forums_12 []byte = []byte(`">`)
var forums_13 []byte = []byte(`</a>
		</span>`)
var forums_14 []byte = []byte(`

		<span style="float: right;">
			<a href="/topic/`)
var forums_15 []byte = []byte(`" style="float: right;font-size: 14px;">`)
var forums_16 []byte = []byte(`</a>
			`)
var forums_17 []byte = []byte(`<br /><span class="rowsmall">`)
var forums_18 []byte = []byte(`</span>`)
var forums_19 []byte = []byte(`
		</span>
		<div style="clear: both;"></div>
	</div>
	`)
var forums_20 []byte = []byte(`<div class="rowitem passive">You don't have access to any forums.</div>`)
var forums_21 []byte = []byte(`
</div>
`)
var topics_0 []byte = []byte(`
<div class="rowblock">
	<div class="rowitem rowhead"><a>Topic List</a></div>
</div>
<div id="topic_list" class="rowblock topic_list">
	`)
var topics_1 []byte = []byte(`<div class="rowitem passive datarow" style="`)
var topics_2 []byte = []byte(`background-image: url(`)
var topics_3 []byte = []byte(`);background-position: left;background-repeat: no-repeat;background-size: 64px;padding-left: 72px;`)
var topics_4 []byte = []byte(`background-color: #FFFFCC;`)
var topics_5 []byte = []byte(`background-color: #eaeaea;`)
var topics_6 []byte = []byte(`">
		<span class="rowsmall" style="float: right;">
			<span class="replyCount">`)
var topics_7 []byte = []byte(` replies</span><br />
			<span class="lastReplyAt">`)
var topics_8 []byte = []byte(`</span>
		</span>
		<span>
			<a class="rowtopic" href="/topic/`)
var topics_9 []byte = []byte(`">`)
var topics_10 []byte = []byte(`</a> `)
var topics_11 []byte = []byte(`<a class="rowsmall" href="/forum/`)
var topics_12 []byte = []byte(`">`)
var topics_13 []byte = []byte(`</a>`)
var topics_14 []byte = []byte(`
			<br /><a class="rowsmall" href="/user/`)
var topics_15 []byte = []byte(`">Starter: `)
var topics_16 []byte = []byte(`</a>
			`)
var topics_17 []byte = []byte(`<span class="rowsmall topic_status_e topic_status_closed" title="Status: Closed"> | &#x1F512;&#xFE0E`)
var topics_18 []byte = []byte(`</span>
		</span>
	</div>
	`)
var topics_19 []byte = []byte(`<div class="rowitem passive">There aren't any topics yet.`)
var topics_20 []byte = []byte(` <a href="/topics/create/">Start one?</a>`)
var topics_21 []byte = []byte(`</div>`)
var topics_22 []byte = []byte(`
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
var forum_10 []byte = []byte(`<span class='username head_tag_upshift' title='No Permissions'>&#x1F512;&#xFE0E</span>`)
var forum_11 []byte = []byte(`<a href="/topics/create/`)
var forum_12 []byte = []byte(`" class='username head_tag_upshift'>New Topic</a>`)
var forum_13 []byte = []byte(`</div>
</div>
<div id="forum_topic_list" class="rowblock topic_list">
	`)
var forum_14 []byte = []byte(`<div class="rowitem passive datarow" style="`)
var forum_15 []byte = []byte(`background-image: url(`)
var forum_16 []byte = []byte(`);background-position: left;background-repeat: no-repeat;background-size: 64px;padding-left: 72px;`)
var forum_17 []byte = []byte(`background-color: #FFFFCC;`)
var forum_18 []byte = []byte(`background-color: #eaeaea;`)
var forum_19 []byte = []byte(`">
		<span class="rowsmall" style="float: right;">
			<span class="replyCount">`)
var forum_20 []byte = []byte(` replies</span><br />
			<span class="lastReplyAt">`)
var forum_21 []byte = []byte(`</span>
		</span>
		<span>
			<a class="rowtopic" href="/topic/`)
var forum_22 []byte = []byte(`">`)
var forum_23 []byte = []byte(`</a>
			<br /><a class="rowsmall" href="/user/`)
var forum_24 []byte = []byte(`">Starter: `)
var forum_25 []byte = []byte(`</a>
			`)
var forum_26 []byte = []byte(`<span class="rowsmall topic_status_e topic_status_closed" title="Status: Closed"> | &#x1F512;&#xFE0E`)
var forum_27 []byte = []byte(`</span>
		</span>
	</div>
	`)
var forum_28 []byte = []byte(`<div class="rowitem passive">There aren't any topics in this forum yet.`)
var forum_29 []byte = []byte(` <a href="/topics/create/`)
var forum_30 []byte = []byte(`">Start one?</a>`)
var forum_31 []byte = []byte(`</div>`)
var forum_32 []byte = []byte(`
</div>
`)
