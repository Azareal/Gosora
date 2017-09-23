package main

// nolint
var header_0 = []byte(`<!doctype html>
<html lang="en">
	<head>
		<title>`)
var header_1 = []byte(`</title>
		<link href="/static/`)
var header_2 = []byte(`/main.css" rel="stylesheet" type="text/css">
		`)
var header_3 = []byte(`
		<link href="/static/`)
var header_4 = []byte(`" rel="stylesheet" type="text/css">
		`)
var header_5 = []byte(`
		<script type="text/javascript" src="/static/jquery-3.1.1.min.js"></script>
		`)
var header_6 = []byte(`
		<script type="text/javascript" src="/static/`)
var header_7 = []byte(`"></script>
		`)
var header_8 = []byte(`
		<script type="text/javascript">var session = "`)
var header_9 = []byte(`";</script>
		<script type="text/javascript" src="/static/global.js"></script>
		<meta name="viewport" content="width=device-width,initial-scale = 1.0, maximum-scale=1.0,user-scalable=no" />
	</head>
	<body>
		<style>`)
var header_10 = []byte(`.supermod_only { display: none !important; }`)
var header_11 = []byte(`</style>
		<div class="container">
`)
var menu_0 = []byte(`<nav class="nav">
	<div class="move_left">
	<div class="move_right">
	<ul>`)
var menu_1 = []byte(`
		<li class="menu_left menu_overview"><a href="/" rel="home">`)
var menu_2 = []byte(`</a></li>
		<li class="menu_left menu_forums"><a href="/forums/">Forums</a></li>
		<li class="menu_left menu_topics"><a href="/">Topics</a></li>
		<li class="menu_left menu_create_topic"><a href="/topics/create/">Create Topic</a></li>
		<li id="general_alerts" class="menu_right menu_alerts">
			<div class="alert_bell"></div>
			<div class="alert_counter"></div>
			<div class="alert_aftercounter"></div>
			<div class="alertList"></div>
		</li>
		`)
var menu_3 = []byte(`
		<li class="menu_left menu_account"><a href="/user/edit/critical/">Account</a></li>
		<li class="menu_left menu_profile"><a href="`)
var menu_4 = []byte(`">Profile</a></li>
		<li class="menu_left menu_account supermod_only"><a href="/panel/">Panel</a></li>
		<li class="menu_left menu_logout"><a href="/accounts/logout/?session=`)
var menu_5 = []byte(`">Logout</a></li>
		`)
var menu_6 = []byte(`
		<li class="menu_left menu_register"><a href="/accounts/create/">Register</a></li>
		<li class="menu_left menu_login"><a href="/accounts/login/">Login</a></li>
		`)
var menu_7 = []byte(`
	</ul>
	</div>
	</div>
	<div style="clear: both;"></div>
</nav>
`)
var header_12 = []byte(`
<div id="back"><div id="main" `)
var header_13 = []byte(`class="shrink_main"`)
var header_14 = []byte(`>
`)
var header_15 = []byte(`<div class="alert">`)
var header_16 = []byte(`</div>`)
var topic_0 = []byte(`

<form id="edit_topic_form" action='/topic/edit/submit/`)
var topic_1 = []byte(`' method="post"></form>
`)
var topic_2 = []byte(`<div id="prevFloat" class="prev_button"><a class="prev_link" aria-label="Go to the previous page" rel="prev" href="/topic/`)
var topic_3 = []byte(`?page=`)
var topic_4 = []byte(`">&lt;</a></div>`)
var topic_5 = []byte(`<link rel="prerender" href="/topic/`)
var topic_6 = []byte(`?page=`)
var topic_7 = []byte(`" />
<div id="nextFloat" class="next_button">
	<a class="next_link" aria-label="Go to the next page" rel="next" href="/topic/`)
var topic_8 = []byte(`?page=`)
var topic_9 = []byte(`">&gt;</a>
</div>`)
var topic_10 = []byte(`

<main>

<div class="rowblock rowhead topic_block">
	<div class="rowitem topic_item`)
var topic_11 = []byte(` topic_sticky_head`)
var topic_12 = []byte(` topic_closed_head`)
var topic_13 = []byte(`">
		<h1 class='topic_name hide_on_edit'>`)
var topic_14 = []byte(`</h1>
		`)
var topic_15 = []byte(`<span class='username hide_on_micro topic_status_e topic_status_closed hide_on_edit' title='Status: Closed'>&#x1F512;&#xFE0E</span>`)
var topic_16 = []byte(`
		<input form='edit_topic_form' class='show_on_edit topic_name_input' name="topic_name" value='`)
var topic_17 = []byte(`' type="text" />
		<button form='edit_topic_form' name="topic-button" class="formbutton show_on_edit submit_edit">Update</button>
		`)
var topic_18 = []byte(`
	</div>
</div>

<article class="rowblock post_container top_post">
	<div class="rowitem passive editable_parent post_item `)
var topic_19 = []byte(`" style="`)
var topic_20 = []byte(`background-image:url(`)
var topic_21 = []byte(`), url(/static/post-avatar-bg.jpg);background-position: 0px `)
var topic_22 = []byte(`-1`)
var topic_23 = []byte(`0px;background-repeat:no-repeat, repeat-y;`)
var topic_24 = []byte(`">
		<p class="hide_on_edit topic_content user_content" style="margin:0;padding:0;">`)
var topic_25 = []byte(`</p>
		<textarea name="topic_content" class="show_on_edit topic_content_input">`)
var topic_26 = []byte(`</textarea>

		<span class="controls">

		<a href="`)
var topic_27 = []byte(`" class="username real_username">`)
var topic_28 = []byte(`</a>&nbsp;&nbsp;
		`)
var topic_29 = []byte(`<a href="/topic/like/submit/`)
var topic_30 = []byte(`" class="mod_button" title="Love it" style="color:#202020;">
		<button class="username like_label"`)
var topic_31 = []byte(` style="background-color:#D6FFD6;"`)
var topic_32 = []byte(`></button></a>`)
var topic_33 = []byte(`<a href='/topic/edit/`)
var topic_34 = []byte(`' class="mod_button open_edit" style="font-weight:normal;" title="Edit Topic"><button class="username edit_label"></button></a>`)
var topic_35 = []byte(`<a href='/topic/delete/submit/`)
var topic_36 = []byte(`' class="mod_button" style="font-weight:normal;" title="Delete Topic"><button class="username trash_label"></button></a>`)
var topic_37 = []byte(`<a class="mod_button" href='/topic/unlock/submit/`)
var topic_38 = []byte(`' style="font-weight:normal;" title="Unlock Topic"><button class="username unlock_label"></button></a>`)
var topic_39 = []byte(`<a href='/topic/lock/submit/`)
var topic_40 = []byte(`' class="mod_button" style="font-weight:normal;" title="Lock Topic"><button class="username lock_label"></button></a>`)
var topic_41 = []byte(`<a class="mod_button" href='/topic/unstick/submit/`)
var topic_42 = []byte(`' style="font-weight:normal;" title="Unpin Topic"><button class="username unpin_label"></button></a>`)
var topic_43 = []byte(`<a href='/topic/stick/submit/`)
var topic_44 = []byte(`' class="mod_button" style="font-weight:normal;" title="Pin Topic"><button class="username pin_label"></button></a>`)
var topic_45 = []byte(`<a class="mod_button" href='/users/ips/?ip=`)
var topic_46 = []byte(`' style="font-weight:normal;" title="View IP"><button class="username ip_label"></button></a>`)
var topic_47 = []byte(`
		<a href="/report/submit/`)
var topic_48 = []byte(`?session=`)
var topic_49 = []byte(`&type=topic" class="mod_button report_item" style="font-weight:normal;" title="Flag Topic"><button class="username flag_label"></button></a>

		`)
var topic_50 = []byte(`<a class="username hide_on_micro like_count">`)
var topic_51 = []byte(`</a><a class="username hide_on_micro like_count_label" title="Like Count"></a>`)
var topic_52 = []byte(`<a class="username hide_on_micro user_tag">`)
var topic_53 = []byte(`</a>`)
var topic_54 = []byte(`<a class="username hide_on_micro level">`)
var topic_55 = []byte(`</a><a class="username hide_on_micro level_label" style="float:right;" title="Level"></a>`)
var topic_56 = []byte(`

		</span>
	</div>
</article>
<div class="rowblock post_container" style="overflow: hidden;">`)
var topic_57 = []byte(`
	<article class="rowitem passive deletable_block editable_parent post_item action_item">
		<span class="action_icon" style="font-size: 18px;padding-right: 5px;">`)
var topic_58 = []byte(`</span>
		<span>`)
var topic_59 = []byte(`</span>
	</article>
`)
var topic_60 = []byte(`
	<article class="rowitem passive deletable_block editable_parent post_item `)
var topic_61 = []byte(`" style="`)
var topic_62 = []byte(`background-image:url(`)
var topic_63 = []byte(`), url(/static/post-avatar-bg.jpg);background-position: 0px `)
var topic_64 = []byte(`-1`)
var topic_65 = []byte(`0px;background-repeat:no-repeat, repeat-y;`)
var topic_66 = []byte(`">
		<p class="editable_block user_content" style="margin:0;padding:0;">`)
var topic_67 = []byte(`</p>

		<span class="controls">

		<a href="`)
var topic_68 = []byte(`" class="username real_username">`)
var topic_69 = []byte(`</a>&nbsp;&nbsp;
		`)
var topic_70 = []byte(`<a href="/reply/like/submit/`)
var topic_71 = []byte(`" class="mod_button" title="Love it" style="color:#202020;"><button class="username like_label"`)
var topic_72 = []byte(` style="background-color:#D6FFD6;"`)
var topic_73 = []byte(`></button></a>`)
var topic_74 = []byte(`<a href="/reply/edit/submit/`)
var topic_75 = []byte(`" class="mod_button" title="Edit Reply"><button class="username edit_item edit_label"></button></a>`)
var topic_76 = []byte(`<a href="/reply/delete/submit/`)
var topic_77 = []byte(`" class="mod_button" title="Delete Reply"><button class="username delete_item trash_label"></button></a>`)
var topic_78 = []byte(`<a class="mod_button" href='/users/ips/?ip=`)
var topic_79 = []byte(`' style="font-weight:normal;" title="View IP"><button class="username ip_label"></button></a>`)
var topic_80 = []byte(`
		<a href="/report/submit/`)
var topic_81 = []byte(`?session=`)
var topic_82 = []byte(`&type=reply" class="mod_button report_item" title="Flag Reply"><button class="username report_item flag_label"></button></a>

		`)
var topic_83 = []byte(`<a class="username hide_on_micro like_count">`)
var topic_84 = []byte(`</a><a class="username hide_on_micro like_count_label" title="Like Count"></a>`)
var topic_85 = []byte(`<a class="username hide_on_micro user_tag">`)
var topic_86 = []byte(`</a>`)
var topic_87 = []byte(`<a class="username hide_on_micro level">`)
var topic_88 = []byte(`</a><a class="username hide_on_micro level_label" style="float:right;" title="Level"></a>`)
var topic_89 = []byte(`

		</span>
	</article>
`)
var topic_90 = []byte(`</div>

`)
var topic_91 = []byte(`
<div class="rowblock topic_reply_form">
	<form action="/reply/create/" method="post">
		<input name="tid" value='`)
var topic_92 = []byte(`' type="hidden" />
		<div class="formrow real_first_child">
			<div class="formitem"><textarea name="reply-content" placeholder="Insert reply here" required></textarea></div>
		</div>
		<div class="formrow">
			<div class="formitem"><button name="reply-button" class="formbutton">Create Reply</button></div>
		</div>
	</form>
</div>
`)
var topic_93 = []byte(`

</main>

`)
var footer_0 = []byte(`<div class="footer">
	<div id="poweredBy">Powered by Gosora - <span>Made with love by Azareal</span></div>
	<form action="/theme/" method="post">
		<div id="themeSelector" style="float: right;">
			<select id="themeSelectorSelect" name="themeSelector" aria-label="Change the site's appearance">
				`)
var footer_1 = []byte(`<option val="`)
var footer_2 = []byte(`"`)
var footer_3 = []byte(` selected`)
var footer_4 = []byte(`>`)
var footer_5 = []byte(`</option>`)
var footer_6 = []byte(`
			</select>
		</div>
	</form>
</div>
					</div>
				`)
var footer_7 = []byte(`<aside class="sidebar">`)
var footer_8 = []byte(`</aside>`)
var footer_9 = []byte(`
				<div style="clear: both;"></div>
			</div>
		</div>
	</body>
</html>
`)
var topic_alt_0 = []byte(`<div id="prevFloat" class="prev_button"><a class="prev_link" aria-label="Go to the previous page" rel="prev" href="/topic/`)
var topic_alt_1 = []byte(`?page=`)
var topic_alt_2 = []byte(`">&lt;</a></div>`)
var topic_alt_3 = []byte(`<link rel="prerender" href="/topic/`)
var topic_alt_4 = []byte(`?page=`)
var topic_alt_5 = []byte(`" />
<div id="nextFloat" class="next_button"><a class="next_link" aria-label="Go to the next page" rel="next" href="/topic/`)
var topic_alt_6 = []byte(`?page=`)
var topic_alt_7 = []byte(`">&gt;</a></div>`)
var topic_alt_8 = []byte(`

<main>

<div class="rowblock rowhead topic_block">
	<form action='/topic/edit/submit/`)
var topic_alt_9 = []byte(`' method="post">
		<div class="rowitem topic_item`)
var topic_alt_10 = []byte(` topic_sticky_head`)
var topic_alt_11 = []byte(` topic_closed_head`)
var topic_alt_12 = []byte(`">
			<h1 class='topic_name hide_on_edit'>`)
var topic_alt_13 = []byte(`</h1>
			`)
var topic_alt_14 = []byte(`<span class='username hide_on_micro topic_status_e topic_status_closed hide_on_edit' title='Status: Closed' style="font-weight:normal;float: right;position:relative;top:-5px;">&#x1F512;&#xFE0E</span>`)
var topic_alt_15 = []byte(`
			<input class='show_on_edit topic_name_input' name="topic_name" value='`)
var topic_alt_16 = []byte(`' type="text" />
			<button name="topic-button" class="formbutton show_on_edit submit_edit">Update</button>
			`)
var topic_alt_17 = []byte(`
		</div>
	</form>
</div>

<div class="rowblock post_container">
	<article class="rowitem passive deletable_block editable_parent post_item top_post">
		<div class="userinfo">
			<div class="avatar_item" style="background-image: url(`)
var topic_alt_18 = []byte(`), url(/static/white-dot.jpg);background-position: 0px -10px;">&nbsp;</div>
			<a href="`)
var topic_alt_19 = []byte(`" class="the_name">`)
var topic_alt_20 = []byte(`</a>
			`)
var topic_alt_21 = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag">`)
var topic_alt_22 = []byte(`</div><div class="tag_post"></div></div>`)
var topic_alt_23 = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag post_level">Level `)
var topic_alt_24 = []byte(`</div><div class="tag_post"></div></div>`)
var topic_alt_25 = []byte(`
		</div>
		<div class="content_container">
			<div class="hide_on_edit topic_content user_content">`)
var topic_alt_26 = []byte(`</div>
			<textarea name="topic_content" class="show_on_edit topic_content_input">`)
var topic_alt_27 = []byte(`</textarea>
			<div class="button_container">
				`)
var topic_alt_28 = []byte(`<a href="/topic/like/submit/`)
var topic_alt_29 = []byte(`" class="action_button">+1</a>`)
var topic_alt_30 = []byte(`<a href="/topic/edit/`)
var topic_alt_31 = []byte(`" class="action_button open_edit">Edit</a>`)
var topic_alt_32 = []byte(`<a href="/topic/delete/submit/`)
var topic_alt_33 = []byte(`" class="action_button delete_item">Delete</a>`)
var topic_alt_34 = []byte(`<a href='/topic/unlock/submit/`)
var topic_alt_35 = []byte(`' class="action_button">Unlock</a>`)
var topic_alt_36 = []byte(`<a href='/topic/lock/submit/`)
var topic_alt_37 = []byte(`' class="action_button">Lock</a>`)
var topic_alt_38 = []byte(`<a href='/topic/unstick/submit/`)
var topic_alt_39 = []byte(`' class="action_button">Unpin</a>`)
var topic_alt_40 = []byte(`<a href='/topic/stick/submit/`)
var topic_alt_41 = []byte(`' class="action_button">Pin</a>`)
var topic_alt_42 = []byte(`
					<a href="/report/submit/`)
var topic_alt_43 = []byte(`?session=`)
var topic_alt_44 = []byte(`&type=topic" class="action_button report_item">Report</a>
				`)
var topic_alt_45 = []byte(`<a class="action_button action_button_right like_count hide_on_micro">`)
var topic_alt_46 = []byte(` up</a>`)
var topic_alt_47 = []byte(`
				<a class="action_button action_button_right created_at hide_on_mobile">`)
var topic_alt_48 = []byte(`</a>
				`)
var topic_alt_49 = []byte(`<a href="#" title="IP Address" class="action_button action_button_right ip_item hide_on_mobile">`)
var topic_alt_50 = []byte(`</a>`)
var topic_alt_51 = []byte(`
			</div>
		</div><div style="clear:both;"></div>
	</article>
	`)
var topic_alt_52 = []byte(`
	<article class="rowitem passive deletable_block editable_parent post_item `)
var topic_alt_53 = []byte(`action_item`)
var topic_alt_54 = []byte(`">
		<div class="userinfo">
			<div class="avatar_item" style="background-image: url(`)
var topic_alt_55 = []byte(`), url(/static/white-dot.jpg);background-position: 0px -10px;">&nbsp;</div>
			<a href="`)
var topic_alt_56 = []byte(`" class="the_name">`)
var topic_alt_57 = []byte(`</a>
			`)
var topic_alt_58 = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag">`)
var topic_alt_59 = []byte(`</div><div class="tag_post"></div></div>`)
var topic_alt_60 = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag post_level">Level `)
var topic_alt_61 = []byte(`</div><div class="tag_post"></div></div>`)
var topic_alt_62 = []byte(`
		</div>
		<div class="content_container" `)
var topic_alt_63 = []byte(`style="margin-left: 0px;"`)
var topic_alt_64 = []byte(`>
			`)
var topic_alt_65 = []byte(`
				<span class="action_icon" style="font-size: 18px;padding-right: 5px;">`)
var topic_alt_66 = []byte(`</span>
				<span>`)
var topic_alt_67 = []byte(`</span>
			`)
var topic_alt_68 = []byte(`
			<div class="editable_block user_content">`)
var topic_alt_69 = []byte(`</div>
			<div class="button_container">
				`)
var topic_alt_70 = []byte(`<a href="/reply/like/submit/`)
var topic_alt_71 = []byte(`" class="action_button">+1</a>`)
var topic_alt_72 = []byte(`<a href="/reply/edit/submit/`)
var topic_alt_73 = []byte(`" class="action_button edit_item">Edit</a>`)
var topic_alt_74 = []byte(`<a href="/reply/delete/submit/`)
var topic_alt_75 = []byte(`" class="action_button delete_item">Delete</a>`)
var topic_alt_76 = []byte(`
					<a href="/report/submit/`)
var topic_alt_77 = []byte(`?session=`)
var topic_alt_78 = []byte(`&type=reply" class="action_button report_item">Report</a>
				`)
var topic_alt_79 = []byte(`<a class="action_button action_button_right like_count hide_on_micro">`)
var topic_alt_80 = []byte(` up</a>`)
var topic_alt_81 = []byte(`
				<a class="action_button action_button_right created_at hide_on_mobile">`)
var topic_alt_82 = []byte(`</a>
				`)
var topic_alt_83 = []byte(`<a href="#" title="IP Address" class="action_button action_button_right ip_item hide_on_mobile">`)
var topic_alt_84 = []byte(`</a>`)
var topic_alt_85 = []byte(`
			</div>
			`)
var topic_alt_86 = []byte(`
		</div>
		<div style="clear:both;"></div>
	</article>
`)
var topic_alt_87 = []byte(`</div>
`)
var topic_alt_88 = []byte(`
<div class="rowblock topic_reply_form">
	<form action="/reply/create/" method="post">
		<input name="tid" value='`)
var topic_alt_89 = []byte(`' type="hidden" />
		<div class="formrow">
			<div class="formitem"><textarea name="reply-content" placeholder="Insert reply here" required></textarea></div>
		</div>
		<div class="formrow">
			<div class="formitem"><button name="reply-button" class="formbutton">Create Reply</button></div>
		</div>
	</form>
</div>
`)
var topic_alt_90 = []byte(`

</main>

`)
var profile_0 = []byte(`

<div id="profile_left_lane" class="colstack_left">
	<!--<header class="colstack_item colstack_head rowhead">
		<div class="rowitem"><h1>Profile</h1></div>
	</header>-->
	<div id="profile_left_pane" class="rowmenu">
		<div class="rowitem avatarRow" style="padding: 0;">
			<img src="`)
var profile_1 = []byte(`" class="avatar" />
		</div>
		<div class="rowitem">`)
var profile_2 = []byte(`
			<span class="profileName">`)
var profile_3 = []byte(`</span>`)
var profile_4 = []byte(`<span class="username" style="float: right;font-weight: normal;">`)
var profile_5 = []byte(`</span>`)
var profile_6 = []byte(`
		</div>
		<div class="rowitem passive">
			<a class="profile_menu_item">Add Friend</a>
		</div>
		`)
var profile_7 = []byte(`<div class="rowitem passive">
			`)
var profile_8 = []byte(`<a href="/users/unban/`)
var profile_9 = []byte(`?session=`)
var profile_10 = []byte(`" class="profile_menu_item">Unban</a>
		`)
var profile_11 = []byte(`<a href="#ban_user" class="profile_menu_item">Ban</a>`)
var profile_12 = []byte(`
		</div>`)
var profile_13 = []byte(`
		<div class="rowitem passive">
			<a href="/report/submit/`)
var profile_14 = []byte(`?session=`)
var profile_15 = []byte(`&type=user" class="profile_menu_item report_item">Report</a>
		</div>
	</div>
</div>

<div id="profile_right_lane" class="colstack_right">
	`)
var profile_16 = []byte(`
	<!-- TODO: Inline the display: none; CSS -->
	<div id="ban_user_head" class="colstack_item colstack_head hash_hide ban_user_hash" style="display: none;">
			<div class="rowitem"><h1>Ban User</h1></div>
	</div>
	<form id="ban_user_form" class="hash_hide ban_user_hash" action="/users/ban/submit/`)
var profile_17 = []byte(`?session=`)
var profile_18 = []byte(`" method="post" style="display: none;">
		`)
var profile_19 = []byte(`
		<div class="colline">If all the fields are left blank, the ban will be permanent.</div>
		<div class="colstack_item">
			<div class="formrow real_first_child">
				<div class="formitem formlabel"><a>Days</a></div>
				<div class="formitem">
					<input name="ban-duration-days" type="number" value="0" min="0" />
				</div>
			</div>
			<div class="formrow">
				<div class="formitem formlabel"><a>Weeks</a></div>
				<div class="formitem">
					<input name="ban-duration-weeks" type="number" value="0" min="0" />
				</div>
			</div>
			<div class="formrow">
				<div class="formitem formlabel"><a>Months</a></div>
				<div class="formitem">
					<input name="ban-duration-months" type="number" value="0" min="0" />
				</div>
			</div>
			<!--<div class="formrow">
				<div class="formitem formlabel"><a>Reason</a></div>
				<div class="formitem"><textarea name="ban-reason" placeholder="A really horrible person" required></textarea></div>
			</div>-->
			<div class="formrow">
				<div class="formitem"><button name="ban-button" class="formbutton form_middle_button">Ban User</button></div>
			</div>
		</div>
	</form>
	`)
var profile_20 = []byte(`

	<div id="profile_comments_head" class="colstack_item colstack_head hash_hide">
		<div class="rowitem"><h1>Comments</h1></div>
	</div>
	<div id="profile_comments" class="colstack_item hash_hide">`)
var profile_21 = []byte(`
		<div class="rowitem passive deletable_block editable_parent simple `)
var profile_22 = []byte(`" style="`)
var profile_23 = []byte(`background-image: url(`)
var profile_24 = []byte(`), url(/static/post-avatar-bg.jpg);background-position: 0px `)
var profile_25 = []byte(`-1`)
var profile_26 = []byte(`0px;`)
var profile_27 = []byte(`">
			<span class="editable_block user_content simple">`)
var profile_28 = []byte(`</span>
			<span class="controls">
				<a href="`)
var profile_29 = []byte(`" class="real_username username">`)
var profile_30 = []byte(`</a>&nbsp;&nbsp;

				`)
var profile_31 = []byte(`<a href="/profile/reply/edit/submit/`)
var profile_32 = []byte(`" class="mod_button" title="Edit Item"><button class="username edit_item edit_label"></button></a>

				<a href="/profile/reply/delete/submit/`)
var profile_33 = []byte(`" class="mod_button" title="Delete Item"><button class="username delete_item trash_label"></button></a>`)
var profile_34 = []byte(`

				<a class="mod_button" href="/report/submit/`)
var profile_35 = []byte(`?session=`)
var profile_36 = []byte(`&type=user-reply"><button class="username report_item flag_label"></button></a>

				`)
var profile_37 = []byte(`<a class="username hide_on_mobile user_tag" style="float: right;">`)
var profile_38 = []byte(`</a>`)
var profile_39 = []byte(`
			</span>
		</div>
	`)
var profile_40 = []byte(`</div>

`)
var profile_41 = []byte(`
	<form id="profile_comments_form" class="hash_hide" action="/profile/reply/create/" method="post">
		<input name="uid" value='`)
var profile_42 = []byte(`' type="hidden" />
		<div class="colstack_item topic_reply_form" style="border-top: none;">
			<div class="formrow">
				<div class="formitem"><textarea name="reply-content" placeholder="Insert reply here"></textarea></div>
			</div>
			<div class="formrow">
				<div class="formitem"><button name="reply-button" class="formbutton">Create Reply</button></div>
			</div>
		</div>
	</form>
`)
var profile_43 = []byte(`
</div>

`)
var profile_44 = []byte(`
<script type="text/javascript">
function handle_profile_hashbit() {
	var hash_class = ""
	switch(window.location.hash.substr(1)) {
		case "ban_user":
			hash_class = "ban_user_hash"
			break
		default:
			console.log("Unknown hashbit")
			return
	}
	$(".hash_hide").hide()
	$("." + hash_class).show()
}
if(window.location.hash) handle_profile_hashbit()
window.addEventListener("hashchange", handle_profile_hashbit, false)
</script>

`)
var forums_0 = []byte(`
<main>

<div class="rowblock opthead">
	<div class="rowitem"><a>Forums</a></div>
</div>
<div class="rowblock">
	`)
var forums_1 = []byte(`<div class="rowitem `)
var forums_2 = []byte(`datarow`)
var forums_3 = []byte(`">
		`)
var forums_4 = []byte(`<span style="float: left;">
			<a href="`)
var forums_5 = []byte(`" style="">`)
var forums_6 = []byte(`</a>
			<br /><span class="rowsmall">`)
var forums_7 = []byte(`</span>
		</span>`)
var forums_8 = []byte(`<span style="float: left;">
			<a href="`)
var forums_9 = []byte(`">`)
var forums_10 = []byte(`</a>
			<br /><span class="rowsmall" style="font-style: italic;">No description</span>
		</span>`)
var forums_11 = []byte(`

		<span style="float: right;">
			<a href="`)
var forums_12 = []byte(`" style="float: right;font-size: 14px;">`)
var forums_13 = []byte(`</a>
			`)
var forums_14 = []byte(`<br /><span class="rowsmall">`)
var forums_15 = []byte(`</span>`)
var forums_16 = []byte(`
		</span>
		<div style="clear: both;"></div>
	</div>
	`)
var forums_17 = []byte(`<div class="rowitem passive">You don't have access to any forums.</div>`)
var forums_18 = []byte(`
</div>

</main>
`)
var topics_0 = []byte(`
<main>

<div class="rowblock rowhead">
	<div class="rowitem"><h1>Topic List</h1></div>
</div>
<div id="topic_list" class="rowblock topic_list">
	`)
var topics_1 = []byte(`<div class="rowitem topic_left passive datarow `)
var topics_2 = []byte(`topic_sticky`)
var topics_3 = []byte(`topic_closed`)
var topics_4 = []byte(`" style="`)
var topics_5 = []byte(`background-image: url(`)
var topics_6 = []byte(`);background-position: left;background-repeat: no-repeat;background-size: 64px;padding-left: 72px;`)
var topics_7 = []byte(`">
		<span class="topic_inner_right rowsmall" style="float: right;">
			<span class="replyCount">`)
var topics_8 = []byte(` replies</span><br />
			<span class="lastReplyAt">`)
var topics_9 = []byte(`</span>
		</span>
		<span>
			<a class="rowtopic" href="`)
var topics_10 = []byte(`">`)
var topics_11 = []byte(`</a> `)
var topics_12 = []byte(`<a class="rowsmall" href="`)
var topics_13 = []byte(`">`)
var topics_14 = []byte(`</a>`)
var topics_15 = []byte(`
			<br /><a class="rowsmall" href="`)
var topics_16 = []byte(`">Starter: `)
var topics_17 = []byte(`</a>
			`)
var topics_18 = []byte(`<span class="rowsmall topic_status_e topic_status_closed" title="Status: Closed"> | &#x1F512;&#xFE0E</span>`)
var topics_19 = []byte(`<span class="rowsmall topic_status_e topic_status_sticky" title="Status: Pinned"> | &#x1F4CD;&#xFE0E</span>`)
var topics_20 = []byte(`
		</span>
	</div>
	<div class="rowitem topic_right passive datarow `)
var topics_21 = []byte(`topic_sticky`)
var topics_22 = []byte(`topic_closed`)
var topics_23 = []byte(`" style="`)
var topics_24 = []byte(`background-image: url(`)
var topics_25 = []byte(`);background-position: left;background-repeat: no-repeat;background-size: 64px;padding-left: 72px;`)
var topics_26 = []byte(`">
		<span>
			<a href="`)
var topics_27 = []byte(`" class="lastName" style="font-size: 14px;">`)
var topics_28 = []byte(`</a><br>
			<span class="rowsmall lastReplyAt">Last: `)
var topics_29 = []byte(`</span>
		</span>
	</div>
	`)
var topics_30 = []byte(`<div class="rowitem passive">There aren't any topics yet.`)
var topics_31 = []byte(` <a href="/topics/create/">Start one?</a>`)
var topics_32 = []byte(`</div>`)
var topics_33 = []byte(`
</div>

</main>
`)
var forum_0 = []byte(`<div id="prevFloat" class="prev_button"><a class="prev_link" aria-label="Go to the previous page" rel="prev" href="/forum/`)
var forum_1 = []byte(`?page=`)
var forum_2 = []byte(`">&lt;</a></div>`)
var forum_3 = []byte(`<link rel="prerender" href="/forum/`)
var forum_4 = []byte(`?page=`)
var forum_5 = []byte(`" />
<div id="nextFloat" class="next_button"><a class="next_link" aria-label="Go to the next page" rel="next" href="/forum/`)
var forum_6 = []byte(`?page=`)
var forum_7 = []byte(`">&gt;</a></div>`)
var forum_8 = []byte(`

<main>

<div id="forum_head_block" class="rowblock rowhead">
	<div class="rowitem forum_title`)
var forum_9 = []byte(` has_opt`)
var forum_10 = []byte(`"><h1>`)
var forum_11 = []byte(`</h1>
	</div>
	`)
var forum_12 = []byte(`
		<div class="opt create_topic_opt" title="Create Topic"><a href="/topics/create/`)
var forum_13 = []byte(`"></a></div>
		`)
var forum_14 = []byte(`<div class="opt locked_opt" title="You don't have the permissions needed to create a topic"><a></a></div>`)
var forum_15 = []byte(`
		<div style="clear: both;"></div>
	`)
var forum_16 = []byte(`
</div>
<div id="forum_topic_list" class="rowblock topic_list">
	`)
var forum_17 = []byte(`<div class="rowitem topic_left passive datarow `)
var forum_18 = []byte(`topic_sticky`)
var forum_19 = []byte(`topic_closed`)
var forum_20 = []byte(`" style="`)
var forum_21 = []byte(`background-image: url(`)
var forum_22 = []byte(`);background-position: left;background-repeat: no-repeat;background-size: 64px;padding-left: 72px;`)
var forum_23 = []byte(`">
		<span class="topic_inner_right rowsmall" style="float: right;">
			<span class="replyCount">`)
var forum_24 = []byte(` replies</span><br />
			<span class="lastReplyAt">`)
var forum_25 = []byte(`</span>
		</span>
		<span>
			<a class="rowtopic" href="`)
var forum_26 = []byte(`">`)
var forum_27 = []byte(`</a>
			<br /><a class="rowsmall" href="`)
var forum_28 = []byte(`">Starter: `)
var forum_29 = []byte(`</a>
			`)
var forum_30 = []byte(`<span class="rowsmall topic_status_e topic_status_closed" title="Status: Closed"> | &#x1F512;&#xFE0E</span>`)
var forum_31 = []byte(`<span class="rowsmall topic_status_e topic_status_sticky" title="Status: Pinned"> | &#x1F4CD;&#xFE0E</span>`)
var forum_32 = []byte(`
		</span>
	</div>
	<div class="rowitem topic_right passive datarow" style="`)
var forum_33 = []byte(`background-image: url(`)
var forum_34 = []byte(`);background-position: left;background-repeat: no-repeat;background-size: 64px;padding-left: 72px;`)
var forum_35 = []byte(`">
		<span>
			<a href="`)
var forum_36 = []byte(`" class="lastName" style="font-size: 14px;">`)
var forum_37 = []byte(`</a><br>
			<span class="rowsmall lastReplyAt">Last: `)
var forum_38 = []byte(`</span>
		</span>
	</div>
	`)
var forum_39 = []byte(`<div class="rowitem passive">There aren't any topics in this forum yet.`)
var forum_40 = []byte(` <a href="/topics/create/`)
var forum_41 = []byte(`">Start one?</a>`)
var forum_42 = []byte(`</div>`)
var forum_43 = []byte(`
</div>

</main>
`)
