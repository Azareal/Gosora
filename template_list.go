package main

// nolint
var header_0 = []byte(`<!doctype html>
<html lang="en">
	<head>
		<title>`)
var header_1 = []byte(` | `)
var header_2 = []byte(`</title>
		<link href="/static/`)
var header_3 = []byte(`/main.css" rel="stylesheet" type="text/css">
		`)
var header_4 = []byte(`
		<link href="/static/`)
var header_5 = []byte(`" rel="stylesheet" type="text/css">
		`)
var header_6 = []byte(`
		<script type="text/javascript" src="/static/jquery-3.1.1.min.js"></script>
		<script type="text/javascript" src="/static/chartist/chartist.min.js"></script>
		`)
var header_7 = []byte(`
		<script type="text/javascript" src="/static/`)
var header_8 = []byte(`"></script>
		`)
var header_9 = []byte(`
		<script type="text/javascript">
		var session = "`)
var header_10 = []byte(`";
		var siteURL = "`)
var header_11 = []byte(`";
		</script>
		<script type="text/javascript" src="/static/global.js"></script>
		<meta name="viewport" content="width=device-width,initial-scale = 1.0, maximum-scale=1.0,user-scalable=no" />
		`)
var header_12 = []byte(`<meta name="description" content="`)
var header_13 = []byte(`" />`)
var header_14 = []byte(`
	</head>
	<body>
		<style>`)
var header_15 = []byte(`.supermod_only { display: none !important; }`)
var header_16 = []byte(`</style>
		<div class="container">
`)
var menu_0 = []byte(`<nav class="nav">
	<div class="move_left">
	<div class="move_right">
	<ul>`)
var menu_1 = []byte(`
		<li id="menu_overview" class="menu_left"><a href="/" rel="home">`)
var menu_2 = []byte(`</a></li>
		<li id="menu_forums" class="menu_left"><a href="/forums/" aria-label="The Forum list" title="Forum List"></a></li>
		<li class="menu_left menu_topics"><a href="/" aria-label="The topic list" title="Topic List"></a></li>
		<li id="general_alerts" class="menu_right menu_alerts">
			<div class="alert_bell"></div>
			<div class="alert_counter" aria-label="The number of alerts"></div>
			<div class="alert_aftercounter"></div>
			<div class="alertList" aria-label="The alert list"></div>
		</li>
		`)
var menu_3 = []byte(`
		<li class="menu_left menu_account"><a href="/user/edit/critical/" aria-label="The account manager" title="Account Manager"></a></li>
		<li class="menu_left menu_profile"><a href="`)
var menu_4 = []byte(`" aria-label="Your profile" title="Your profile"></a></li>
		<li class="menu_left menu_panel menu_account supermod_only"><a href="/panel/" aria-label="The Control Panel" title="Control Panel"></a></li>
		<li class="menu_left menu_logout"><a href="/accounts/logout/?session=`)
var menu_5 = []byte(`" aria-label="Log out of your account" title="Logout"></a></li>
		`)
var menu_6 = []byte(`
		<li class="menu_left menu_register"><a href="/accounts/create/" aria-label="Create a new account" title="Register"></a></li>
		<li class="menu_left menu_login"><a href="/accounts/login/" aria-label="Login to your account" title="Login"></a></li>
		`)
var menu_7 = []byte(`
		<li class="menu_left menu_hamburger" title="Menu"><a></a></li>
	</ul>
	</div>
	</div>
	<div style="clear: both;"></div>
</nav>
`)
var header_17 = []byte(`
<div id="back"><div id="main" `)
var header_18 = []byte(`class="shrink_main"`)
var header_19 = []byte(`>
`)
var header_20 = []byte(`<div class="alert">`)
var header_21 = []byte(`</div>`)
var topic_0 = []byte(`

<form id="edit_topic_form" action='/topic/edit/submit/`)
var topic_1 = []byte(`?session=`)
var topic_2 = []byte(`' method="post"></form>
`)
var topic_3 = []byte(`<link rel="prev" href="/topic/`)
var topic_4 = []byte(`?page=`)
var topic_5 = []byte(`" />
<div id="prevFloat" class="prev_button"><a class="prev_link" aria-label="Go to the previous page" rel="prev" href="/topic/`)
var topic_6 = []byte(`?page=`)
var topic_7 = []byte(`">&lt;</a></div>`)
var topic_8 = []byte(`<link rel="prerender next" href="/topic/`)
var topic_9 = []byte(`?page=`)
var topic_10 = []byte(`" />
<div id="nextFloat" class="next_button">
	<a class="next_link" aria-label="Go to the next page" rel="next" href="/topic/`)
var topic_11 = []byte(`?page=`)
var topic_12 = []byte(`">&gt;</a>
</div>`)
var topic_13 = []byte(`

<main>

<div class="rowblock rowhead topic_block" aria-label="The opening post of this topic">
	<div class="rowitem topic_item`)
var topic_14 = []byte(` topic_sticky_head`)
var topic_15 = []byte(` topic_closed_head`)
var topic_16 = []byte(`">
		<h1 class='topic_name hide_on_edit'>`)
var topic_17 = []byte(`</h1>
		`)
var topic_18 = []byte(`<span class='username hide_on_micro topic_status_e topic_status_closed hide_on_edit' title='Status: Closed'>&#x1F512;&#xFE0E</span>`)
var topic_19 = []byte(`
		<input form='edit_topic_form' class='show_on_edit topic_name_input' name="topic_name" value='`)
var topic_20 = []byte(`' type="text" />
		<button form='edit_topic_form' name="topic-button" class="formbutton show_on_edit submit_edit">Update</button>
		`)
var topic_21 = []byte(`
	</div>
</div>
`)
var topic_22 = []byte(`
<article class="rowblock post_container poll">
	<div class="rowitem passive editable_parent post_item poll_item `)
var topic_23 = []byte(`" style="background-image: url(`)
var topic_24 = []byte(`), url(/static/`)
var topic_25 = []byte(`/post-avatar-bg.jpg);background-position: 0px `)
var topic_26 = []byte(`-1`)
var topic_27 = []byte(`0px;background-repeat:no-repeat, repeat-y;">
		<div class="topic_content user_content" style="margin:0;padding:0;">
			`)
var topic_28 = []byte(`
			<div class="poll_option">
				<input form="poll_`)
var topic_29 = []byte(`_form" id="poll_option_`)
var topic_30 = []byte(`" name="poll_option_input" type="checkbox" value="`)
var topic_31 = []byte(`" />
				<label class="poll_option_label" for="poll_option_`)
var topic_32 = []byte(`">
					<div class="sel"></div>
				</label>
				<span id="poll_option_text_`)
var topic_33 = []byte(`" class="poll_option_text">`)
var topic_34 = []byte(`</span>
			</div>
			`)
var topic_35 = []byte(`
			<div class="poll_buttons">
				<button form="poll_`)
var topic_36 = []byte(`_form" class="poll_vote_button">Vote</button>
				<button class="poll_results_button" data-poll-id="`)
var topic_37 = []byte(`">Results</button>
				<a href="#"><button class="poll_cancel_button">Cancel</button></a>
			</div>
		</div>
		<div id="poll_results_`)
var topic_38 = []byte(`" class="poll_results auto_hide">
			<div class="topic_content user_content"></div>
		</div>
	</div>
</article>
`)
var topic_39 = []byte(`

<article itemscope itemtype="http://schema.org/CreativeWork" class="rowblock post_container top_post" aria-label="The opening post for this topic">
	<div class="rowitem passive editable_parent post_item `)
var topic_40 = []byte(`" style="background-image: url(`)
var topic_41 = []byte(`), url(/static/`)
var topic_42 = []byte(`/post-avatar-bg.jpg);background-position: 0px `)
var topic_43 = []byte(`-1`)
var topic_44 = []byte(`0px;background-repeat:no-repeat, repeat-y;">
		<p class="hide_on_edit topic_content user_content" itemprop="text" style="margin:0;padding:0;">`)
var topic_45 = []byte(`</p>
		<textarea name="topic_content" class="show_on_edit topic_content_input">`)
var topic_46 = []byte(`</textarea>

		<span class="controls" aria-label="Controls and Author Information">

		<a href="`)
var topic_47 = []byte(`" class="username real_username" rel="author">`)
var topic_48 = []byte(`</a>&nbsp;&nbsp;
		`)
var topic_49 = []byte(`<a href="/topic/like/submit/`)
var topic_50 = []byte(`?session=`)
var topic_51 = []byte(`" class="mod_button" title="Love it" `)
var topic_52 = []byte(`aria-label="Unlike this topic"`)
var topic_53 = []byte(`aria-label="Like this topic"`)
var topic_54 = []byte(` style="color:#202020;">
		<button class="username like_label"`)
var topic_55 = []byte(` style="background-color:#D6FFD6;"`)
var topic_56 = []byte(`></button></a>`)
var topic_57 = []byte(`<a href='/topic/edit/`)
var topic_58 = []byte(`' class="mod_button open_edit" style="font-weight:normal;" title="Edit Topic" aria-label="Edit this topic"><button class="username edit_label"></button></a>`)
var topic_59 = []byte(`<a href='/topic/delete/submit/`)
var topic_60 = []byte(`?session=`)
var topic_61 = []byte(`' class="mod_button" style="font-weight:normal;" title="Delete Topic" aria-label="Delete this topic"><button class="username trash_label"></button></a>`)
var topic_62 = []byte(`<a class="mod_button" href='/topic/unlock/submit/`)
var topic_63 = []byte(`?session=`)
var topic_64 = []byte(`' style="font-weight:normal;" title="Unlock Topic" aria-label="Unlock this topic"><button class="username unlock_label"></button></a>`)
var topic_65 = []byte(`<a href='/topic/lock/submit/`)
var topic_66 = []byte(`?session=`)
var topic_67 = []byte(`' class="mod_button" style="font-weight:normal;" title="Lock Topic" aria-label="Lock this topic"><button class="username lock_label"></button></a>`)
var topic_68 = []byte(`<a class="mod_button" href='/topic/unstick/submit/`)
var topic_69 = []byte(`?session=`)
var topic_70 = []byte(`' style="font-weight:normal;" title="Unpin Topic" aria-label="Unpin this topic"><button class="username unpin_label"></button></a>`)
var topic_71 = []byte(`<a href='/topic/stick/submit/`)
var topic_72 = []byte(`?session=`)
var topic_73 = []byte(`' class="mod_button" style="font-weight:normal;" title="Pin Topic" aria-label="Pin this topic"><button class="username pin_label"></button></a>`)
var topic_74 = []byte(`<a class="mod_button" href='/users/ips/?ip=`)
var topic_75 = []byte(`' style="font-weight:normal;" title="View IP" aria-label="The poster's IP is `)
var topic_76 = []byte(`"><button class="username ip_label"></button></a>`)
var topic_77 = []byte(`
		<a href="/report/submit/`)
var topic_78 = []byte(`?session=`)
var topic_79 = []byte(`&type=topic" class="mod_button report_item" style="font-weight:normal;" title="Flag this topic" aria-label="Flag this topic" rel="nofollow"><button class="username flag_label"></button></a>

		`)
var topic_80 = []byte(`<a class="username hide_on_micro like_count" aria-label="The number of likes on this topic">`)
var topic_81 = []byte(`</a><a class="username hide_on_micro like_count_label" title="Like Count"></a>`)
var topic_82 = []byte(`<a class="username hide_on_micro user_tag">`)
var topic_83 = []byte(`</a>`)
var topic_84 = []byte(`<a class="username hide_on_micro level" aria-label="The poster's level">`)
var topic_85 = []byte(`</a><a class="username hide_on_micro level_label" style="float:right;" title="Level"></a>`)
var topic_86 = []byte(`

		</span>
	</div>
</article>

<div class="rowblock post_container" aria-label="The current page for this topic" style="overflow: hidden;">`)
var topic_87 = []byte(`
	<article itemscope itemtype="http://schema.org/CreativeWork" class="rowitem passive deletable_block editable_parent post_item action_item">
		<span class="action_icon" style="font-size: 18px;padding-right: 5px;">`)
var topic_88 = []byte(`</span>
		<span itemprop="text">`)
var topic_89 = []byte(`</span>
	</article>
`)
var topic_90 = []byte(`
	<article itemscope itemtype="http://schema.org/CreativeWork" class="rowitem passive deletable_block editable_parent post_item `)
var topic_91 = []byte(`" style="background-image: url(`)
var topic_92 = []byte(`), url(/static/`)
var topic_93 = []byte(`/post-avatar-bg.jpg);background-position: 0px `)
var topic_94 = []byte(`-1`)
var topic_95 = []byte(`0px;background-repeat:no-repeat, repeat-y;">
		`)
var topic_96 = []byte(`
		<p class="editable_block user_content" itemprop="text" style="margin:0;padding:0;">`)
var topic_97 = []byte(`</p>

		<span class="controls">

		<a href="`)
var topic_98 = []byte(`" class="username real_username" rel="author">`)
var topic_99 = []byte(`</a>&nbsp;&nbsp;
		`)
var topic_100 = []byte(`<a href="/reply/like/submit/`)
var topic_101 = []byte(`?session=`)
var topic_102 = []byte(`" class="mod_button" title="Love it" style="color:#202020;"><button class="username like_label"`)
var topic_103 = []byte(` style="background-color:#D6FFD6;"`)
var topic_104 = []byte(`></button></a>`)
var topic_105 = []byte(`<a href="/reply/edit/submit/`)
var topic_106 = []byte(`?session=`)
var topic_107 = []byte(`" class="mod_button" title="Edit Reply"><button class="username edit_item edit_label"></button></a>`)
var topic_108 = []byte(`<a href="/reply/delete/submit/`)
var topic_109 = []byte(`?session=`)
var topic_110 = []byte(`" class="mod_button" title="Delete Reply"><button class="username delete_item trash_label"></button></a>`)
var topic_111 = []byte(`<a class="mod_button" href='/users/ips/?ip=`)
var topic_112 = []byte(`' style="font-weight:normal;" title="View IP"><button class="username ip_label"></button></a>`)
var topic_113 = []byte(`
		<a href="/report/submit/`)
var topic_114 = []byte(`?session=`)
var topic_115 = []byte(`&type=reply" class="mod_button report_item" title="Flag this reply" aria-label="Flag this reply" rel="nofollow"><button class="username report_item flag_label"></button></a>

		`)
var topic_116 = []byte(`<a class="username hide_on_micro like_count">`)
var topic_117 = []byte(`</a><a class="username hide_on_micro like_count_label" title="Like Count"></a>`)
var topic_118 = []byte(`<a class="username hide_on_micro user_tag">`)
var topic_119 = []byte(`</a>`)
var topic_120 = []byte(`<a class="username hide_on_micro level">`)
var topic_121 = []byte(`</a><a class="username hide_on_micro level_label" style="float:right;" title="Level"></a>`)
var topic_122 = []byte(`

		</span>
	</article>
`)
var topic_123 = []byte(`</div>

`)
var topic_124 = []byte(`
<div class="rowblock topic_reply_form quick_create_form">
	<form id="quick_post_form" enctype="multipart/form-data" action="/reply/create/?session=`)
var topic_125 = []byte(`" method="post"></form>
	<input form="quick_post_form" name="tid" value='`)
var topic_126 = []byte(`' type="hidden" />
	<input form="quick_post_form" id="has_poll_input" name="has_poll" value="0" type="hidden" />
	<div class="formrow real_first_child">
		<div class="formitem">
			<textarea id="input_content" form="quick_post_form" name="reply-content" placeholder="Insert reply here" required></textarea>
		</div>
	</div>
	<div class="formrow poll_content_row auto_hide">
		<div class="formitem">
			<div class="pollinput" data-pollinput="0">
				<input type="checkbox" disabled />
				<label class="pollinputlabel"></label>
				<input form="quick_post_form" name="pollinputitem[0]" class="pollinputinput" type="text" placeholder="Add new poll option" />
			</div>
		</div>
	</div>
	<div class="formrow quick_button_row">
		<div class="formitem">
			<button form="quick_post_form" name="reply-button" class="formbutton">Create Reply</button>
			<button form="quick_post_form" class="formbutton" id="add_poll_button">Add Poll</button>
			`)
var topic_127 = []byte(`
			<input name="upload_files" form="quick_post_form" id="upload_files" multiple type="file" style="display: none;" />
			<label for="upload_files" class="formbutton add_file_button">Add File</label>
			<div id="upload_file_dock"></div>`)
var topic_128 = []byte(`
		</div>
	</div>
</div>
`)
var topic_129 = []byte(`

</main>

`)
var footer_0 = []byte(`<div class="footer">
	`)
var footer_1 = []byte(`
	<div id="poweredByHolder" class="footerBit">
		<div id="poweredBy">
			<a id="poweredByName" href="https://github.com/Azareal/Gosora">Powered by Gosora</a><span id="poweredByDash"> - </span><span id="poweredByMaker">Made with love by Azareal</span>
		</div>
		<form action="/theme/" method="post">
			<div id="themeSelector" style="float: right;">
				<select id="themeSelectorSelect" name="themeSelector" aria-label="Change the site's appearance">
				`)
var footer_2 = []byte(`<option val="`)
var footer_3 = []byte(`"`)
var footer_4 = []byte(` selected`)
var footer_5 = []byte(`>`)
var footer_6 = []byte(`</option>`)
var footer_7 = []byte(`
				</select>
			</div>
		</form>
	</div>
</div>
					</div>
				<aside class="sidebar">`)
var footer_8 = []byte(`</aside>
				<div style="clear: both;"></div>
			</div>
		</div>
	</body>
</html>
`)
var topic_alt_0 = []byte(`<link rel="prev" href="/topic/`)
var topic_alt_1 = []byte(`?page=`)
var topic_alt_2 = []byte(`" />
<div id="prevFloat" class="prev_button"><a class="prev_link" aria-label="Go to the previous page" rel="prev" href="/topic/`)
var topic_alt_3 = []byte(`?page=`)
var topic_alt_4 = []byte(`">&lt;</a></div>`)
var topic_alt_5 = []byte(`<link rel="prerender next" href="/topic/`)
var topic_alt_6 = []byte(`?page=`)
var topic_alt_7 = []byte(`" />
<div id="nextFloat" class="next_button"><a class="next_link" aria-label="Go to the next page" rel="next" href="/topic/`)
var topic_alt_8 = []byte(`?page=`)
var topic_alt_9 = []byte(`">&gt;</a></div>`)
var topic_alt_10 = []byte(`

<main>

<div class="rowblock rowhead topic_block" aria-label="The opening post of this topic">
	<form action='/topic/edit/submit/`)
var topic_alt_11 = []byte(`?session=`)
var topic_alt_12 = []byte(`' method="post">
		<div class="rowitem topic_item`)
var topic_alt_13 = []byte(` topic_sticky_head`)
var topic_alt_14 = []byte(` topic_closed_head`)
var topic_alt_15 = []byte(`">
			<h1 class='topic_name hide_on_edit'>`)
var topic_alt_16 = []byte(`</h1>
			`)
var topic_alt_17 = []byte(`<span class='username hide_on_micro topic_status_e topic_status_closed hide_on_edit' title='Status: Closed' style="font-weight:normal;float: right;position:relative;top:-5px;">&#x1F512;&#xFE0E</span>`)
var topic_alt_18 = []byte(`
			<input class='show_on_edit topic_name_input' name="topic_name" value='`)
var topic_alt_19 = []byte(`' type="text" />
			<button name="topic-button" class="formbutton show_on_edit submit_edit">Update</button>
			`)
var topic_alt_20 = []byte(`
		</div>
	</form>
</div>

<div class="rowblock post_container">
	`)
var topic_alt_21 = []byte(`
	<form id="poll_`)
var topic_alt_22 = []byte(`_form" action="/poll/vote/`)
var topic_alt_23 = []byte(`?session=`)
var topic_alt_24 = []byte(`" method="post"></form>
	<article class="rowitem passive deletable_block editable_parent post_item poll_item top_post hide_on_edit">
		<div class="userinfo" aria-label="The information on the poster">
			<div class="avatar_item" style="background-image: url(`)
var topic_alt_25 = []byte(`), url(/static/white-dot.jpg);background-position: 0px -10px;">&nbsp;</div>
			<a href="`)
var topic_alt_26 = []byte(`" class="the_name" rel="author">`)
var topic_alt_27 = []byte(`</a>
			`)
var topic_alt_28 = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag">`)
var topic_alt_29 = []byte(`</div><div class="tag_post"></div></div>`)
var topic_alt_30 = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag post_level">Level `)
var topic_alt_31 = []byte(`</div><div class="tag_post"></div></div>`)
var topic_alt_32 = []byte(`
		</div>
		<div id="poll_voter_`)
var topic_alt_33 = []byte(`" class="content_container poll_voter">
			<div class="topic_content user_content">
				`)
var topic_alt_34 = []byte(`
				<div class="poll_option">
					<input form="poll_`)
var topic_alt_35 = []byte(`_form" id="poll_option_`)
var topic_alt_36 = []byte(`" name="poll_option_input" type="checkbox" value="`)
var topic_alt_37 = []byte(`" />
					<label class="poll_option_label" for="poll_option_`)
var topic_alt_38 = []byte(`">
						<div class="sel"></div>
					</label>
					<span id="poll_option_text_`)
var topic_alt_39 = []byte(`" class="poll_option_text">`)
var topic_alt_40 = []byte(`</span>
				</div>
				`)
var topic_alt_41 = []byte(`
				<div class="poll_buttons">
					<button form="poll_`)
var topic_alt_42 = []byte(`_form" class="poll_vote_button">Vote</button>
					<button class="poll_results_button" data-poll-id="`)
var topic_alt_43 = []byte(`">Results</button>
					<a href="#"><button class="poll_cancel_button">Cancel</button></a>
				</div>
			</div>
		</div>
		<div id="poll_results_`)
var topic_alt_44 = []byte(`" class="content_container poll_results auto_hide">
			<div class="topic_content user_content"></div>
		</div>
	</article>
	`)
var topic_alt_45 = []byte(`
	<article itemscope itemtype="http://schema.org/CreativeWork" class="rowitem passive deletable_block editable_parent post_item top_post" aria-label="The opening post for this topic">
		<div class="userinfo" aria-label="The information on the poster">
			<div class="avatar_item" style="background-image: url(`)
var topic_alt_46 = []byte(`), url(/static/white-dot.jpg);background-position: 0px -10px;">&nbsp;</div>
			<a href="`)
var topic_alt_47 = []byte(`" class="the_name" rel="author">`)
var topic_alt_48 = []byte(`</a>
			`)
var topic_alt_49 = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag">`)
var topic_alt_50 = []byte(`</div><div class="tag_post"></div></div>`)
var topic_alt_51 = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag post_level">Level `)
var topic_alt_52 = []byte(`</div><div class="tag_post"></div></div>`)
var topic_alt_53 = []byte(`
		</div>
		<div class="content_container">
			<div class="hide_on_edit topic_content user_content" itemprop="text">`)
var topic_alt_54 = []byte(`</div>
			<textarea name="topic_content" class="show_on_edit topic_content_input">`)
var topic_alt_55 = []byte(`</textarea>
			<div class="button_container">
				`)
var topic_alt_56 = []byte(`<a href="/topic/like/submit/`)
var topic_alt_57 = []byte(`?session=`)
var topic_alt_58 = []byte(`" class="action_button like_item add_like" aria-label="Like this post" data-action="like"></a>`)
var topic_alt_59 = []byte(`<a href="/topic/edit/`)
var topic_alt_60 = []byte(`" class="action_button open_edit" aria-label="Edit this post" data-action="edit"></a>`)
var topic_alt_61 = []byte(`<a href="/topic/delete/submit/`)
var topic_alt_62 = []byte(`?session=`)
var topic_alt_63 = []byte(`" class="action_button delete_item" aria-label="Delete this post" data-action="delete"></a>`)
var topic_alt_64 = []byte(`<a href='/topic/unlock/submit/`)
var topic_alt_65 = []byte(`?session=`)
var topic_alt_66 = []byte(`' class="action_button unlock_item" data-action="unlock"></a>`)
var topic_alt_67 = []byte(`<a href='/topic/lock/submit/`)
var topic_alt_68 = []byte(`?session=`)
var topic_alt_69 = []byte(`' class="action_button lock_item" data-action="lock"></a>`)
var topic_alt_70 = []byte(`<a href='/topic/unstick/submit/`)
var topic_alt_71 = []byte(`?session=`)
var topic_alt_72 = []byte(`' class="action_button unpin_item" data-action="unpin"></a>`)
var topic_alt_73 = []byte(`<a href='/topic/stick/submit/`)
var topic_alt_74 = []byte(`?session=`)
var topic_alt_75 = []byte(`' class="action_button pin_item" data-action="pin"></a>`)
var topic_alt_76 = []byte(`<a href="/users/ips/?ip=`)
var topic_alt_77 = []byte(`" title="IP Address" class="action_button ip_item_button hide_on_big" aria-label="This user's IP" data-action="ip"></a>`)
var topic_alt_78 = []byte(`
					<a href="/report/submit/`)
var topic_alt_79 = []byte(`?session=`)
var topic_alt_80 = []byte(`&type=topic" class="action_button report_item" aria-label="Report this post" data-action="report"></a>
					<a href="#" class="action_button button_menu"></a>
				`)
var topic_alt_81 = []byte(`
				<div class="action_button_right`)
var topic_alt_82 = []byte(` has_likes`)
var topic_alt_83 = []byte(`">
					`)
var topic_alt_84 = []byte(`<a class="action_button like_count hide_on_micro">`)
var topic_alt_85 = []byte(`</a>`)
var topic_alt_86 = []byte(`
					<a class="action_button created_at hide_on_mobile">`)
var topic_alt_87 = []byte(`</a>
					`)
var topic_alt_88 = []byte(`<a href="/users/ips/?ip=`)
var topic_alt_89 = []byte(`" title="IP Address" class="action_button ip_item hide_on_mobile">`)
var topic_alt_90 = []byte(`</a>`)
var topic_alt_91 = []byte(`
				</div>
			</div>
		</div><div style="clear:both;"></div>
	</article>

	`)
var topic_alt_92 = []byte(`
	<article itemscope itemtype="http://schema.org/CreativeWork" class="rowitem passive deletable_block editable_parent post_item `)
var topic_alt_93 = []byte(`action_item`)
var topic_alt_94 = []byte(`">
		<div class="userinfo" aria-label="The information on the poster">
			<div class="avatar_item" style="background-image: url(`)
var topic_alt_95 = []byte(`), url(/static/white-dot.jpg);background-position: 0px -10px;">&nbsp;</div>
			<a href="`)
var topic_alt_96 = []byte(`" class="the_name" rel="author">`)
var topic_alt_97 = []byte(`</a>
			`)
var topic_alt_98 = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag">`)
var topic_alt_99 = []byte(`</div><div class="tag_post"></div></div>`)
var topic_alt_100 = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag post_level">Level `)
var topic_alt_101 = []byte(`</div><div class="tag_post"></div></div>`)
var topic_alt_102 = []byte(`
		</div>
		<div class="content_container" `)
var topic_alt_103 = []byte(`style="margin-left: 0px;"`)
var topic_alt_104 = []byte(`>
			`)
var topic_alt_105 = []byte(`
				<span class="action_icon" style="font-size: 18px;padding-right: 5px;">`)
var topic_alt_106 = []byte(`</span>
				<span itemprop="text">`)
var topic_alt_107 = []byte(`</span>
			`)
var topic_alt_108 = []byte(`
			<div class="editable_block user_content" itemprop="text">`)
var topic_alt_109 = []byte(`</div>
			<div class="button_container">
				`)
var topic_alt_110 = []byte(`<a href="/reply/like/submit/`)
var topic_alt_111 = []byte(`?session=`)
var topic_alt_112 = []byte(`" class="action_button like_item add_like" aria-label="Like this post" data-action="like"></a>`)
var topic_alt_113 = []byte(`<a href="/reply/edit/submit/`)
var topic_alt_114 = []byte(`?session=`)
var topic_alt_115 = []byte(`" class="action_button edit_item" aria-label="Edit this post" data-action="edit"></a>`)
var topic_alt_116 = []byte(`<a href="/reply/delete/submit/`)
var topic_alt_117 = []byte(`?session=`)
var topic_alt_118 = []byte(`" class="action_button delete_item" aria-label="Delete this post" data-action="delete"></a>`)
var topic_alt_119 = []byte(`<a href="/users/ips/?ip=`)
var topic_alt_120 = []byte(`" title="IP Address" class="action_button ip_item_button hide_on_big" aria-label="This user's IP Address" data-action="ip"></a>`)
var topic_alt_121 = []byte(`
					<a href="/report/submit/`)
var topic_alt_122 = []byte(`?session=`)
var topic_alt_123 = []byte(`&type=reply" class="action_button report_item" aria-label="Report this post" data-action="report"></a>
					<a href="#" class="action_button button_menu"></a>
				`)
var topic_alt_124 = []byte(`
				<div class="action_button_right`)
var topic_alt_125 = []byte(` has_likes`)
var topic_alt_126 = []byte(`">
					`)
var topic_alt_127 = []byte(`<a class="action_button like_count hide_on_micro">`)
var topic_alt_128 = []byte(`</a>`)
var topic_alt_129 = []byte(`
					<a class="action_button created_at hide_on_mobile">`)
var topic_alt_130 = []byte(`</a>
					`)
var topic_alt_131 = []byte(`<a href="/users/ips/?ip=`)
var topic_alt_132 = []byte(`" title="IP Address" class="action_button ip_item hide_on_mobile">`)
var topic_alt_133 = []byte(`</a>`)
var topic_alt_134 = []byte(`
				</div>
			</div>
			`)
var topic_alt_135 = []byte(`
		</div>
		<div style="clear:both;"></div>
	</article>
`)
var topic_alt_136 = []byte(`</div>

`)
var topic_alt_137 = []byte(`
<div class="rowblock topic_reply_container">
	<div class="userinfo" aria-label="The information on the poster">
		<div class="avatar_item" style="background-image: url(`)
var topic_alt_138 = []byte(`), url(/static/white-dot.jpg);background-position: 0px -10px;">&nbsp;</div>
		<a href="`)
var topic_alt_139 = []byte(`" class="the_name" rel="author">`)
var topic_alt_140 = []byte(`</a>
		`)
var topic_alt_141 = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag">`)
var topic_alt_142 = []byte(`</div><div class="tag_post"></div></div>`)
var topic_alt_143 = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag post_level">Level `)
var topic_alt_144 = []byte(`</div><div class="tag_post"></div></div>`)
var topic_alt_145 = []byte(`
	</div>
	<div class="rowblock topic_reply_form quick_create_form">
		<form id="quick_post_form" enctype="multipart/form-data" action="/reply/create/?session=`)
var topic_alt_146 = []byte(`" method="post"></form>
		<input form="quick_post_form" name="tid" value='`)
var topic_alt_147 = []byte(`' type="hidden" />
		<input form="quick_post_form" id="has_poll_input" name="has_poll" value="0" type="hidden" />
		<div class="formrow real_first_child">
			<div class="formitem">
				<textarea id="input_content" form="quick_post_form" name="reply-content" placeholder="What do you think?" required></textarea>
			</div>
		</div>
		<div class="formrow poll_content_row auto_hide">
			<div class="formitem">
				<div class="pollinput" data-pollinput="0">
					<input type="checkbox" disabled />
					<label class="pollinputlabel"></label>
					<input form="quick_post_form" name="pollinputitem[0]" class="pollinputinput" type="text" placeholder="Add new poll option" />
				</div>
			</div>
		</div>
		<div class="formrow quick_button_row">
			<div class="formitem">
				<button form="quick_post_form" name="reply-button" class="formbutton">Create Reply</button>
				<button form="quick_post_form" class="formbutton" id="add_poll_button">Add Poll</button>
				`)
var topic_alt_148 = []byte(`
				<input name="upload_files" form="quick_post_form" id="upload_files" multiple type="file" style="display: none;" />
				<label for="upload_files" class="formbutton add_file_button">Add File</label>
				<div id="upload_file_dock"></div>`)
var topic_alt_149 = []byte(`
			</div>
		</div>
	</div>
</div>
`)
var topic_alt_150 = []byte(`

</main>

`)
var profile_0 = []byte(`

<div id="profile_container" class="colstack">

<div id="profile_left_lane" class="colstack_left">
	<div id="profile_left_pane" class="rowmenu">
		<div class="topBlock">
			<div class="rowitem avatarRow">
				<img src="`)
var profile_1 = []byte(`" class="avatar" alt="`)
var profile_2 = []byte(`'s Avatar" title="`)
var profile_3 = []byte(`'s Avatar" />
			</div>
			<div class="rowitem nameRow">
				<span class="profileName">`)
var profile_4 = []byte(`</span>`)
var profile_5 = []byte(`<span class="username">`)
var profile_6 = []byte(`</span>`)
var profile_7 = []byte(`
			</div>
		</div>
		<div class="passiveBlock">
			<div class="rowitem passive">
				<a class="profile_menu_item">Add Friend</a>
			</div>
			`)
var profile_8 = []byte(`<div class="rowitem passive">
				`)
var profile_9 = []byte(`<a href="/users/unban/`)
var profile_10 = []byte(`?session=`)
var profile_11 = []byte(`" class="profile_menu_item">Unban</a>
			`)
var profile_12 = []byte(`<a href="#ban_user" class="profile_menu_item">Ban</a>`)
var profile_13 = []byte(`
			</div>`)
var profile_14 = []byte(`
			<div class="rowitem passive">
				<a href="/report/submit/`)
var profile_15 = []byte(`?session=`)
var profile_16 = []byte(`&type=user" class="profile_menu_item report_item" aria-label="Report User" title="Report User"></a>
			</div>
		</div>
	</div>
</div>

<div id="profile_right_lane" class="colstack_right">
	`)
var profile_17 = []byte(`
	<!-- TODO: Inline the display: none; CSS -->
	<div id="ban_user_head" class="colstack_item colstack_head hash_hide ban_user_hash" style="display: none;">
			<div class="rowitem"><h1><a>Ban User</a></h1></div>
	</div>
	<form id="ban_user_form" class="hash_hide ban_user_hash" action="/users/ban/submit/`)
var profile_18 = []byte(`?session=`)
var profile_19 = []byte(`" method="post" style="display: none;">
		`)
var profile_20 = []byte(`
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
var profile_21 = []byte(`

	<div id="profile_comments_head" class="colstack_item colstack_head hash_hide">
		<div class="rowitem"><h1><a>Comments</a></h1></div>
	</div>
	<div id="profile_comments" class="colstack_item hash_hide">`)
var profile_comments_row_0 = []byte(`
		<div class="rowitem passive deletable_block editable_parent simple `)
var profile_comments_row_1 = []byte(`" style="background-image: url(`)
var profile_comments_row_2 = []byte(`), url(/static/post-avatar-bg.jpg);background-position: 0px `)
var profile_comments_row_3 = []byte(`-1`)
var profile_comments_row_4 = []byte(`0px;">
			<span class="editable_block user_content simple">`)
var profile_comments_row_5 = []byte(`</span>
			<span class="controls">
				<a href="`)
var profile_comments_row_6 = []byte(`" class="real_username username">`)
var profile_comments_row_7 = []byte(`</a>&nbsp;&nbsp;

				`)
var profile_comments_row_8 = []byte(`<a href="/profile/reply/edit/submit/`)
var profile_comments_row_9 = []byte(`?session=`)
var profile_comments_row_10 = []byte(`" class="mod_button" title="Edit Item"><button class="username edit_item edit_label"></button></a>

				<a href="/profile/reply/delete/submit/`)
var profile_comments_row_11 = []byte(`?session=`)
var profile_comments_row_12 = []byte(`" class="mod_button" title="Delete Item"><button class="username delete_item trash_label"></button></a>`)
var profile_comments_row_13 = []byte(`

				<a class="mod_button" href="/report/submit/`)
var profile_comments_row_14 = []byte(`?session=`)
var profile_comments_row_15 = []byte(`&type=user-reply"><button class="username report_item flag_label"></button></a>

				`)
var profile_comments_row_16 = []byte(`<a class="username hide_on_mobile user_tag" style="float: right;">`)
var profile_comments_row_17 = []byte(`</a>`)
var profile_comments_row_18 = []byte(`
			</span>
		</div>
	`)
var profile_comments_row_19 = []byte(`
		<div class="rowitem passive deletable_block editable_parent comment `)
var profile_comments_row_20 = []byte(`">
			<div class="topRow">
				<div class="userbit">
					<img src="`)
var profile_comments_row_21 = []byte(`" alt="`)
var profile_comments_row_22 = []byte(`'s Avatar" title="`)
var profile_comments_row_23 = []byte(`'s Avatar" />
					<span class="nameAndTitle">
						<a href="`)
var profile_comments_row_24 = []byte(`" class="real_username username">`)
var profile_comments_row_25 = []byte(`</a>
						`)
var profile_comments_row_26 = []byte(`<a class="username hide_on_mobile user_tag" style="float: right;">`)
var profile_comments_row_27 = []byte(`</a>`)
var profile_comments_row_28 = []byte(`
					</span>
				</div>
				<span class="controls">
					`)
var profile_comments_row_29 = []byte(`
						<a href="/profile/reply/edit/submit/`)
var profile_comments_row_30 = []byte(`?session=`)
var profile_comments_row_31 = []byte(`" class="mod_button" title="Edit Item"><button class="username edit_item edit_label"></button></a>
						<a href="/profile/reply/delete/submit/`)
var profile_comments_row_32 = []byte(`?session=`)
var profile_comments_row_33 = []byte(`" class="mod_button" title="Delete Item"><button class="username delete_item trash_label"></button></a>
					`)
var profile_comments_row_34 = []byte(`
					<a class="mod_button" href="/report/submit/`)
var profile_comments_row_35 = []byte(`?session=`)
var profile_comments_row_36 = []byte(`&type=user-reply"><button class="username report_item flag_label"></button></a>
				</span>
			</div>
			<div class="content_column">
				<span class="editable_block user_content">`)
var profile_comments_row_37 = []byte(`</span>
			</div>
		</div>
	`)
var profile_22 = []byte(`</div>

`)
var profile_23 = []byte(`
	<form id="profile_comments_form" class="hash_hide" action="/profile/reply/create/?session=`)
var profile_24 = []byte(`" method="post">
		<input name="uid" value='`)
var profile_25 = []byte(`' type="hidden" />
		<div class="colstack_item topic_reply_form" style="border-top: none;">
			<div class="formrow">
				<div class="formitem"><textarea class="input_content" name="reply-content" placeholder="Insert comment here"></textarea></div>
			</div>
			<div class="formrow quick_button_row">
				<div class="formitem"><button name="reply-button" class="formbutton">Create Reply</button></div>
			</div>
		</div>
	</form>
`)
var profile_26 = []byte(`
</div>

</div>

`)
var profile_27 = []byte(`
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
<main itemscope itemtype="http://schema.org/ItemList">

<div class="rowblock opthead">
	<div class="rowitem"><h1 itemprop="name">Forums</h1></div>
</div>
<div class="rowblock forum_list">
	`)
var forums_1 = []byte(`<div class="rowitem `)
var forums_2 = []byte(`datarow `)
var forums_3 = []byte(`"itemprop="itemListElement" itemscope
      itemtype="http://schema.org/ListItem">
		<span class="forum_left shift_left">
			<a href="`)
var forums_4 = []byte(`" itemprop="item">`)
var forums_5 = []byte(`</a>
		`)
var forums_6 = []byte(`
			<br /><span class="rowsmall" itemprop="description">`)
var forums_7 = []byte(`</span>
		`)
var forums_8 = []byte(`
			<br /><span class="rowsmall" style="font-style: italic;">No description</span>
		`)
var forums_9 = []byte(`
		</span>

		<span class="forum_right shift_right">
			`)
var forums_10 = []byte(`<img class="extra_little_row_avatar" src="`)
var forums_11 = []byte(`" height=64 width=64 alt="`)
var forums_12 = []byte(`'s Avatar" title="`)
var forums_13 = []byte(`'s Avatar" />`)
var forums_14 = []byte(`
			<span>
				<a href="`)
var forums_15 = []byte(`">`)
var forums_16 = []byte(`None`)
var forums_17 = []byte(`</a>
				`)
var forums_18 = []byte(`<br /><span class="rowsmall">`)
var forums_19 = []byte(`</span>`)
var forums_20 = []byte(`
			</span>
		</span>
		<div style="clear: both;"></div>
	</div>
	`)
var forums_21 = []byte(`<div class="rowitem passive">You don't have access to any forums.</div>`)
var forums_22 = []byte(`
</div>

</main>
`)
var topics_0 = []byte(`
<main itemscope itemtype="http://schema.org/ItemList">

<div class="rowblock rowhead topic_list_title_block">
	<div class="rowitem topic_list_title`)
var topics_1 = []byte(` has_opt`)
var topics_2 = []byte(`"><h1 itemprop="name">All Topics</h1></div>
	`)
var topics_3 = []byte(`
		<div class="pre_opt auto_hide"></div>
		<div class="opt create_topic_opt" title="Create Topic" aria-label="Create a topic"><a class="create_topic_link" href="/topics/create/"></a></div>
		`)
var topics_4 = []byte(`
		<div class="opt mod_opt" title="Moderate">
			<a class="moderate_link" href="#" aria-label="Moderate Posts"></a>
		</div>
		`)
var topics_5 = []byte(`<div class="opt locked_opt" title="You don't have the permissions needed to create a topic" aria-label="You don't have the permissions needed to make a topic anywhere"><a></a></div>`)
var topics_6 = []byte(`
		<div style="clear: both;"></div>
	`)
var topics_7 = []byte(`
</div>

`)
var topics_8 = []byte(`
<div class="mod_floater auto_hide">
	<form method="post">
	<div class="mod_floater_head">
		<span>What do you want to do with these 18 topics?</span>
	</div>
	<div class="mod_floater_body">
		<select class="mod_floater_options">
			<option val="delete">Delete them</option>
			<option val="lock">Lock them</option>
			<option val="move">Move them</option>
		</select>
		<button class="mod_floater_submit">Run</button>
	</div>
	</form>
</div>

`)
var topics_9 = []byte(`
<div id="mod_topic_mover" class="modal_pane auto_hide">
	<form action="/topic/move/submit/?session=`)
var topics_10 = []byte(`" method="post">
		<input id="mover_fid" name="fid" value="0" type="hidden" />
		<div class="pane_header">
			<h3>Move these topics to?</h3>
		</div>
		<div class="pane_body">
			<div class="pane_table">
				`)
var topics_11 = []byte(`<div id="mover_fid_`)
var topics_12 = []byte(`" data-fid="`)
var topics_13 = []byte(`" class="pane_row">`)
var topics_14 = []byte(`</div>`)
var topics_15 = []byte(`
			</div>
		</div>
		<div class="pane_buttons">
			<button id="mover_submit">Move Topics</button>
		</div>
	</form>
</div>
<div class="rowblock topic_create_form quick_create_form" style="display: none;" aria-label="Quick Topic Form">
	<form name="topic_create_form_form" id="quick_post_form" enctype="multipart/form-data" action="/topic/create/submit/?session=`)
var topics_16 = []byte(`" method="post"></form>
	<input form="quick_post_form" id="has_poll_input" name="has_poll" value="0" type="hidden" />
	<img class="little_row_avatar" src="`)
var topics_17 = []byte(`" height="64" alt="Your Avatar" title="Your Avatar" />
	<div class="main_form">
		<div class="topic_meta">
			<div class="formrow topic_board_row real_first_child">
				<div class="formitem"><select form="quick_post_form" id="topic_board_input" name="topic-board">
					`)
var topics_18 = []byte(`<option `)
var topics_19 = []byte(`selected`)
var topics_20 = []byte(` value="`)
var topics_21 = []byte(`">`)
var topics_22 = []byte(`</option>`)
var topics_23 = []byte(`
				</select></div>
			</div>
			<div class="formrow topic_name_row">
				<div class="formitem">
					<input form="quick_post_form" name="topic-name" placeholder="What's up?" required>
				</div>
			</div>
		</div>
		<div class="formrow topic_content_row">
			<div class="formitem">
				<textarea form="quick_post_form" id="input_content" name="topic-content" placeholder="Insert post here" required></textarea>
			</div>
		</div>
		<div class="formrow poll_content_row auto_hide">
			<div class="formitem">
				<div class="pollinput" data-pollinput="0">
					<input type="checkbox" disabled />
					<label class="pollinputlabel"></label>
					<input form="quick_post_form" name="pollinputitem[0]" class="pollinputinput" type="text" placeholder="Add new poll option" />
				</div>
			</div>
		</div>
		<div class="formrow quick_button_row">
			<div class="formitem">
				<button form="quick_post_form" class="formbutton">Create Topic</button>
				<button form="quick_post_form" class="formbutton" id="add_poll_button">Add Poll</button>
				`)
var topics_24 = []byte(`
				<input name="upload_files" form="quick_post_form" id="upload_files" multiple type="file" style="display: none;" />
				<label for="upload_files" class="formbutton add_file_button">Add File</label>
				<div id="upload_file_dock"></div>`)
var topics_25 = []byte(`
				<button class="formbutton close_form">Cancel</button>
			</div>
		</div>
	</div>
</div>
	`)
var topics_26 = []byte(`
<div id="topic_list" class="rowblock topic_list" aria-label="A list containing topics from every forum">
	`)
var topics_27 = []byte(`<div class="topic_row" data-tid="`)
var topics_28 = []byte(`">
	<div class="rowitem topic_left passive datarow `)
var topics_29 = []byte(`topic_sticky`)
var topics_30 = []byte(`topic_closed`)
var topics_31 = []byte(`">
		<span class="selector"></span>
		<a href="`)
var topics_32 = []byte(`"><img src="`)
var topics_33 = []byte(`" height="64" alt="`)
var topics_34 = []byte(`'s Avatar" title="`)
var topics_35 = []byte(`'s Avatar" /></a>
		<span class="topic_inner_left">
			<a class="rowtopic" href="`)
var topics_36 = []byte(`" itemprop="itemListElement"><span>`)
var topics_37 = []byte(`</span></a> `)
var topics_38 = []byte(`<a class="rowsmall parent_forum" href="`)
var topics_39 = []byte(`">`)
var topics_40 = []byte(`</a>`)
var topics_41 = []byte(`
			<br /><a class="rowsmall starter" href="`)
var topics_42 = []byte(`">`)
var topics_43 = []byte(`</a>
			`)
var topics_44 = []byte(`<span class="rowsmall topic_status_e topic_status_closed" title="Status: Closed"> | &#x1F512;&#xFE0E</span>`)
var topics_45 = []byte(`<span class="rowsmall topic_status_e topic_status_sticky" title="Status: Pinned"> | &#x1F4CD;&#xFE0E</span>`)
var topics_46 = []byte(`
		</span>
		<span class="topic_inner_right rowsmall" style="float: right;">
			<span class="replyCount">`)
var topics_47 = []byte(`</span><br />
			<span class="likeCount">`)
var topics_48 = []byte(`</span>
		</span>
	</div>
	<div class="rowitem topic_right passive datarow `)
var topics_49 = []byte(`topic_sticky`)
var topics_50 = []byte(`topic_closed`)
var topics_51 = []byte(`">
		<a href="`)
var topics_52 = []byte(`"><img src="`)
var topics_53 = []byte(`" height="64" alt="`)
var topics_54 = []byte(`'s Avatar" title="`)
var topics_55 = []byte(`'s Avatar" /></a>
		<span>
			<a href="`)
var topics_56 = []byte(`" class="lastName" style="font-size: 14px;">`)
var topics_57 = []byte(`</a><br>
			<span class="rowsmall lastReplyAt">`)
var topics_58 = []byte(`</span>
		</span>
	</div>
	</div>`)
var topics_59 = []byte(`<div class="rowitem passive">There aren't any topics yet.`)
var topics_60 = []byte(` <a href="/topics/create/">Start one?</a>`)
var topics_61 = []byte(`</div>`)
var topics_62 = []byte(`
</div>

`)
var topics_63 = []byte(`
<div class="pageset">
	`)
var topics_64 = []byte(`<div class="pageitem"><a href="?page=`)
var topics_65 = []byte(`" rel="prev" aria-label="Go to the previous page">Prev</a></div>
	<link rel="prev" href="?page=`)
var topics_66 = []byte(`" />`)
var topics_67 = []byte(`
	<div class="pageitem"><a href="?page=`)
var topics_68 = []byte(`">`)
var topics_69 = []byte(`</a></div>
	`)
var topics_70 = []byte(`
	<link rel="next" href="?page=`)
var topics_71 = []byte(`" />
	<div class="pageitem"><a href="?page=`)
var topics_72 = []byte(`" rel="next" aria-label="Go to the next page">Next</a></div>`)
var topics_73 = []byte(`
</div>
`)
var topics_74 = []byte(`

</main>
`)
var forum_0 = []byte(`<div id="prevFloat" class="prev_button"><a class="prev_link" aria-label="Go to the previous page" rel="prev" href="/forum/`)
var forum_1 = []byte(`?page=`)
var forum_2 = []byte(`">&lt;</a></div>`)
var forum_3 = []byte(`<div id="nextFloat" class="next_button"><a class="next_link" aria-label="Go to the next page" rel="next" href="/forum/`)
var forum_4 = []byte(`?page=`)
var forum_5 = []byte(`">&gt;</a></div>`)
var forum_6 = []byte(`

<main itemscope itemtype="http://schema.org/ItemList">
	<div id="forum_head_block" class="rowblock rowhead topic_list_title_block">
		<div class="rowitem forum_title`)
var forum_7 = []byte(` has_opt`)
var forum_8 = []byte(`">
			<h1 itemprop="name">`)
var forum_9 = []byte(`</h1>
		</div>
		`)
var forum_10 = []byte(`
			<div class="pre_opt auto_hide"></div>
			<div class="opt create_topic_opt" title="Create Topic" aria-label="Create a topic"><a class="create_topic_link" href="/topics/create/`)
var forum_11 = []byte(`"></a></div>
			`)
var forum_12 = []byte(`
			<div class="opt mod_opt" title="Moderate">
				<a class="moderate_link" href="#" aria-label="Moderate Posts"></a>
			</div>
			`)
var forum_13 = []byte(`<div class="opt locked_opt" title="You don't have the permissions needed to create a topic" aria-label="You don't have the permissions needed to make a topic in this forum"><a></a></div>`)
var forum_14 = []byte(`
			<div style="clear: both;"></div>
		`)
var forum_15 = []byte(`
	</div>
	`)
var forum_16 = []byte(`
	<div class="mod_floater auto_hide">
		<form method="post">
			<div class="mod_floater_head">
				<span>What do you want to do with these 18 topics?</span>
			</div>
			<div class="mod_floater_body">
				<select class="mod_floater_options">
					<option val="delete">Delete them</option>
					<option val="lock">Lock them</option>
					<option val="move">Move them</option>
				</select>
				<button>Run</button>
			</div>
		</form>
	</div>
	`)
var forum_17 = []byte(`
	<div id="forum_topic_create_form" class="rowblock topic_create_form quick_create_form" style="display: none;" aria-label="Quick Topic Form">
		<form id="quick_post_form" enctype="multipart/form-data" action="/topic/create/submit/" method="post"></form>
		<img class="little_row_avatar" src="`)
var forum_18 = []byte(`" height="64" alt="Your Avatar" title="Your Avatar" />
		<input form="quick_post_form" id="topic_board_input" name="topic-board" value="`)
var forum_19 = []byte(`" type="hidden">
		<div class="main_form">
			<div class="topic_meta">
				<div class="formrow topic_name_row real_first_child">
					<div class="formitem">
						<input form="quick_post_form" name="topic-name" placeholder="What's up?" required>
					</div>
				</div>
			</div>
			<div class="formrow topic_content_row">
				<div class="formitem">
					<textarea form="quick_post_form" id="input_content" name="topic-content" placeholder="Insert post here" required></textarea>
				</div>
			</div>
			<div class="formrow poll_content_row auto_hide">
				<div class="formitem">
					Poll stuff
				</div>
			</div>
			<div class="formrow quick_button_row">
				<div class="formitem">
					<button form="quick_post_form" name="topic-button" class="formbutton">Create Topic</button>
					<button form="quick_post_form" class="formbutton" id="add_poll_button">Add Poll</button>
					`)
var forum_20 = []byte(`
					<input name="upload_files" form="quick_post_form" id="upload_files" multiple type="file" style="display: none;" />
					<label for="upload_files" class="formbutton add_file_button">Add File</label>
					<div id="upload_file_dock"></div>`)
var forum_21 = []byte(`
					<button class="formbutton close_form">Cancel</button>
				</div>
			</div>
		</div>
	</div>
	`)
var forum_22 = []byte(`
	<div id="forum_topic_list" class="rowblock topic_list">
		`)
var forum_23 = []byte(`<div class="topic_row" data-tid="`)
var forum_24 = []byte(`">
		<div class="rowitem topic_left passive datarow `)
var forum_25 = []byte(`topic_sticky`)
var forum_26 = []byte(`topic_closed`)
var forum_27 = []byte(`">
			<span class="selector"></span>
			<a href="`)
var forum_28 = []byte(`"><img src="`)
var forum_29 = []byte(`" height="64" alt="`)
var forum_30 = []byte(`'s Avatar" title="`)
var forum_31 = []byte(`'s Avatar" /></a>
			<span class="topic_inner_left">
				<a class="rowtopic" href="`)
var forum_32 = []byte(`" itemprop="itemListElement"><span>`)
var forum_33 = []byte(`</span></a>
				<br /><a class="rowsmall starter" href="`)
var forum_34 = []byte(`">`)
var forum_35 = []byte(`</a>
				`)
var forum_36 = []byte(`<span class="rowsmall topic_status_e topic_status_closed" title="Status: Closed"> | &#x1F512;&#xFE0E</span>`)
var forum_37 = []byte(`<span class="rowsmall topic_status_e topic_status_sticky" title="Status: Pinned"> | &#x1F4CD;&#xFE0E</span>`)
var forum_38 = []byte(`
			</span>
			<span class="topic_inner_right rowsmall" style="float: right;">
				<span class="replyCount">`)
var forum_39 = []byte(`</span><br />
				<span class="likeCount">`)
var forum_40 = []byte(`</span>
			</span>
		</div>
		<div class="rowitem topic_right passive datarow `)
var forum_41 = []byte(`topic_sticky`)
var forum_42 = []byte(`topic_closed`)
var forum_43 = []byte(`">
			<a href="`)
var forum_44 = []byte(`"><img src="`)
var forum_45 = []byte(`" height="64" alt="`)
var forum_46 = []byte(`'s Avatar" title="`)
var forum_47 = []byte(`'s Avatar" /></a>
			<span>
				<a href="`)
var forum_48 = []byte(`" class="lastName" style="font-size: 14px;">`)
var forum_49 = []byte(`</a><br>
				<span class="rowsmall lastReplyAt">`)
var forum_50 = []byte(`</span>
			</span>
		</div>
		</div>`)
var forum_51 = []byte(`<div class="rowitem passive">There aren't any topics in this forum yet.`)
var forum_52 = []byte(` <a href="/topics/create/`)
var forum_53 = []byte(`">Start one?</a>`)
var forum_54 = []byte(`</div>`)
var forum_55 = []byte(`
	</div>

`)
var forum_56 = []byte(`
<div class="pageset">
	`)
var forum_57 = []byte(`<div class="pageitem"><a href="?page=`)
var forum_58 = []byte(`" rel="prev" aria-label="Go to the previous page">Prev</a></div>
	<link rel="prev" href="?page=`)
var forum_59 = []byte(`" />`)
var forum_60 = []byte(`
	<div class="pageitem"><a href="?page=`)
var forum_61 = []byte(`">`)
var forum_62 = []byte(`</a></div>
	`)
var forum_63 = []byte(`
	<link rel="next" href="?page=`)
var forum_64 = []byte(`" />
	<div class="pageitem"><a href="?page=`)
var forum_65 = []byte(`" rel="next" aria-label="Go to the next page">Next</a></div>`)
var forum_66 = []byte(`
</div>
`)
var forum_67 = []byte(`

</main>
`)
var guilds_guild_list_0 = []byte(`
<main>
	<div class="rowblock opthead">
		<div class="rowitem"><a>Guild List</a></div>
	</div>
	<div class="rowblock">
		`)
var guilds_guild_list_1 = []byte(`<div class="rowitem datarow">
			<span style="float: left;">
				<a href="`)
var guilds_guild_list_2 = []byte(`" style="">`)
var guilds_guild_list_3 = []byte(`</a>
				<br /><span class="rowsmall">`)
var guilds_guild_list_4 = []byte(`</span>
			</span>
			<span style="float: right;">
				<span style="float: right;font-size: 14px;">`)
var guilds_guild_list_5 = []byte(` members</span>
				<br /><span class="rowsmall">`)
var guilds_guild_list_6 = []byte(`</span>
			</span>
			<div style="clear: both;"></div>
		</div>
		`)
var guilds_guild_list_7 = []byte(`<div class="rowitem passive">There aren't any visible guilds.</div>`)
var guilds_guild_list_8 = []byte(`
	</div>
</main>
`)
