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
		<li id="menu_forums" class="menu_left"><a href="/forums/" aria-label="`)
var menu_3 = []byte(`" title="`)
var menu_4 = []byte(`"></a></li>
		<li class="menu_left menu_topics"><a href="/" aria-label="`)
var menu_5 = []byte(`" title="`)
var menu_6 = []byte(`"></a></li>
		<li id="general_alerts" class="menu_right menu_alerts">
			<div class="alert_bell"></div>
			<div class="alert_counter" aria-label="`)
var menu_7 = []byte(`"></div>
			<div class="alert_aftercounter"></div>
			<div class="alertList" aria-label="`)
var menu_8 = []byte(`"></div>
		</li>
		`)
var menu_9 = []byte(`
		<li class="menu_left menu_account"><a href="/user/edit/critical/" aria-label="`)
var menu_10 = []byte(`" title="`)
var menu_11 = []byte(`"></a></li>
		<li class="menu_left menu_profile"><a href="`)
var menu_12 = []byte(`" aria-label="`)
var menu_13 = []byte(`" title="`)
var menu_14 = []byte(`"></a></li>
		<li class="menu_left menu_panel menu_account supermod_only"><a href="/panel/" aria-label="`)
var menu_15 = []byte(`" title="`)
var menu_16 = []byte(`"></a></li>
		<li class="menu_left menu_logout"><a href="/accounts/logout/?session=`)
var menu_17 = []byte(`" aria-label="`)
var menu_18 = []byte(`" title="`)
var menu_19 = []byte(`"></a></li>
		`)
var menu_20 = []byte(`
		<li class="menu_left menu_register"><a href="/accounts/create/" aria-label="`)
var menu_21 = []byte(`" title="`)
var menu_22 = []byte(`"></a></li>
		<li class="menu_left menu_login"><a href="/accounts/login/" aria-label="`)
var menu_23 = []byte(`" title="`)
var menu_24 = []byte(`"></a></li>
		`)
var menu_25 = []byte(`
		<li class="menu_left menu_hamburger" title="`)
var menu_26 = []byte(`"><a></a></li>
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
<div class="alertbox">`)
var header_20 = []byte(`
	<div class="alert">`)
var header_21 = []byte(`</div>`)
var header_22 = []byte(`
</div>
`)
var topic_0 = []byte(`

<form id="edit_topic_form" action='/topic/edit/submit/`)
var topic_1 = []byte(`?session=`)
var topic_2 = []byte(`' method="post"></form>
`)
var topic_3 = []byte(`<link rel="prev" href="/topic/`)
var topic_4 = []byte(`?page=`)
var topic_5 = []byte(`" />
<div id="prevFloat" class="prev_button"><a class="prev_link" aria-label="`)
var topic_6 = []byte(`" rel="prev" href="/topic/`)
var topic_7 = []byte(`?page=`)
var topic_8 = []byte(`">`)
var topic_9 = []byte(`</a></div>`)
var topic_10 = []byte(`<link rel="prerender next" href="/topic/`)
var topic_11 = []byte(`?page=`)
var topic_12 = []byte(`" />
<div id="nextFloat" class="next_button">
	<a class="next_link" aria-label="`)
var topic_13 = []byte(`" rel="next" href="/topic/`)
var topic_14 = []byte(`?page=`)
var topic_15 = []byte(`">`)
var topic_16 = []byte(`</a>
</div>`)
var topic_17 = []byte(`

<main>

<div class="rowblock rowhead topic_block" aria-label="`)
var topic_18 = []byte(`">
	<div class="rowitem topic_item`)
var topic_19 = []byte(` topic_sticky_head`)
var topic_20 = []byte(` topic_closed_head`)
var topic_21 = []byte(`">
		<h1 class='topic_name hide_on_edit'>`)
var topic_22 = []byte(`</h1>
		`)
var topic_23 = []byte(`<span class='username hide_on_micro topic_status_e topic_status_closed hide_on_edit' title='`)
var topic_24 = []byte(`' aria-label='`)
var topic_25 = []byte(`'>&#x1F512;&#xFE0E</span>`)
var topic_26 = []byte(`
		<input form='edit_topic_form' class='show_on_edit topic_name_input' name="topic_name" value='`)
var topic_27 = []byte(`' type="text" aria-label="`)
var topic_28 = []byte(`" />
		<button form='edit_topic_form' name="topic-button" class="formbutton show_on_edit submit_edit">`)
var topic_29 = []byte(`</button>
		`)
var topic_30 = []byte(`
	</div>
</div>
`)
var topic_31 = []byte(`
<article class="rowblock post_container poll" aria-level="`)
var topic_32 = []byte(`">
	<div class="rowitem passive editable_parent post_item poll_item `)
var topic_33 = []byte(`" style="background-image: url(`)
var topic_34 = []byte(`), url(/static/`)
var topic_35 = []byte(`/post-avatar-bg.jpg);background-position: 0px `)
var topic_36 = []byte(`-1`)
var topic_37 = []byte(`0px;background-repeat:no-repeat, repeat-y;">
		<div class="topic_content user_content" style="margin:0;padding:0;">
			`)
var topic_38 = []byte(`
			<div class="poll_option">
				<input form="poll_`)
var topic_39 = []byte(`_form" id="poll_option_`)
var topic_40 = []byte(`" name="poll_option_input" type="checkbox" value="`)
var topic_41 = []byte(`" />
				<label class="poll_option_label" for="poll_option_`)
var topic_42 = []byte(`">
					<div class="sel"></div>
				</label>
				<span id="poll_option_text_`)
var topic_43 = []byte(`" class="poll_option_text">`)
var topic_44 = []byte(`</span>
			</div>
			`)
var topic_45 = []byte(`
			<div class="poll_buttons">
				<button form="poll_`)
var topic_46 = []byte(`_form" class="poll_vote_button">`)
var topic_47 = []byte(`</button>
				<button class="poll_results_button" data-poll-id="`)
var topic_48 = []byte(`">`)
var topic_49 = []byte(`</button>
				<a href="#"><button class="poll_cancel_button">`)
var topic_50 = []byte(`</button></a>
			</div>
		</div>
		<div id="poll_results_`)
var topic_51 = []byte(`" class="poll_results auto_hide">
			<div class="topic_content user_content"></div>
		</div>
	</div>
</article>
`)
var topic_52 = []byte(`

<article itemscope itemtype="http://schema.org/CreativeWork" class="rowblock post_container top_post" aria-label="`)
var topic_53 = []byte(`">
	<div class="rowitem passive editable_parent post_item `)
var topic_54 = []byte(`" style="background-image: url(`)
var topic_55 = []byte(`), url(/static/`)
var topic_56 = []byte(`/post-avatar-bg.jpg);background-position: 0px `)
var topic_57 = []byte(`-1`)
var topic_58 = []byte(`0px;background-repeat:no-repeat, repeat-y;">
		<p class="hide_on_edit topic_content user_content" itemprop="text" style="margin:0;padding:0;">`)
var topic_59 = []byte(`</p>
		<textarea name="topic_content" class="show_on_edit topic_content_input">`)
var topic_60 = []byte(`</textarea>

		<span class="controls" aria-label="`)
var topic_61 = []byte(`">

		<a href="`)
var topic_62 = []byte(`" class="username real_username" rel="author">`)
var topic_63 = []byte(`</a>&nbsp;&nbsp;
		`)
var topic_64 = []byte(`<a href="/topic/like/submit/`)
var topic_65 = []byte(`?session=`)
var topic_66 = []byte(`" class="mod_button"`)
var topic_67 = []byte(` title="`)
var topic_68 = []byte(`" aria-label="`)
var topic_69 = []byte(`"`)
var topic_70 = []byte(` title="`)
var topic_71 = []byte(`" aria-label="`)
var topic_72 = []byte(`"`)
var topic_73 = []byte(` style="color:#202020;">
		<button class="username like_label"`)
var topic_74 = []byte(` style="background-color:#D6FFD6;"`)
var topic_75 = []byte(`></button></a>`)
var topic_76 = []byte(`<a href='/topic/edit/`)
var topic_77 = []byte(`' class="mod_button open_edit" style="font-weight:normal;" title="`)
var topic_78 = []byte(`" aria-label="`)
var topic_79 = []byte(`"><button class="username edit_label"></button></a>`)
var topic_80 = []byte(`<a href='/topic/delete/submit/`)
var topic_81 = []byte(`?session=`)
var topic_82 = []byte(`' class="mod_button" style="font-weight:normal;" title="`)
var topic_83 = []byte(`" aria-label="`)
var topic_84 = []byte(`"><button class="username trash_label"></button></a>`)
var topic_85 = []byte(`<a class="mod_button" href='/topic/unlock/submit/`)
var topic_86 = []byte(`?session=`)
var topic_87 = []byte(`' style="font-weight:normal;" title="`)
var topic_88 = []byte(`" aria-label="`)
var topic_89 = []byte(`"><button class="username unlock_label"></button></a>`)
var topic_90 = []byte(`<a href='/topic/lock/submit/`)
var topic_91 = []byte(`?session=`)
var topic_92 = []byte(`' class="mod_button" style="font-weight:normal;" title="`)
var topic_93 = []byte(`" aria-label="`)
var topic_94 = []byte(`"><button class="username lock_label"></button></a>`)
var topic_95 = []byte(`<a class="mod_button" href='/topic/unstick/submit/`)
var topic_96 = []byte(`?session=`)
var topic_97 = []byte(`' style="font-weight:normal;" title="`)
var topic_98 = []byte(`" aria-label="`)
var topic_99 = []byte(`"><button class="username unpin_label"></button></a>`)
var topic_100 = []byte(`<a href='/topic/stick/submit/`)
var topic_101 = []byte(`?session=`)
var topic_102 = []byte(`' class="mod_button" style="font-weight:normal;" title="`)
var topic_103 = []byte(`" aria-label="`)
var topic_104 = []byte(`"><button class="username pin_label"></button></a>`)
var topic_105 = []byte(`<a class="mod_button" href='/users/ips/?ip=`)
var topic_106 = []byte(`' style="font-weight:normal;" title="`)
var topic_107 = []byte(`" aria-label="The poster's IP is `)
var topic_108 = []byte(`"><button class="username ip_label"></button></a>`)
var topic_109 = []byte(`
		<a href="/report/submit/`)
var topic_110 = []byte(`?session=`)
var topic_111 = []byte(`&type=topic" class="mod_button report_item" style="font-weight:normal;" title="`)
var topic_112 = []byte(`" aria-label="`)
var topic_113 = []byte(`" rel="nofollow"><button class="username flag_label"></button></a>

		`)
var topic_114 = []byte(`<a class="username hide_on_micro like_count" aria-label="`)
var topic_115 = []byte(`">`)
var topic_116 = []byte(`</a><a class="username hide_on_micro like_count_label" title="`)
var topic_117 = []byte(`"></a>`)
var topic_118 = []byte(`<a class="username hide_on_micro user_tag">`)
var topic_119 = []byte(`</a>`)
var topic_120 = []byte(`<a class="username hide_on_micro level" aria-label="`)
var topic_121 = []byte(`">`)
var topic_122 = []byte(`</a><a class="username hide_on_micro level_label" style="float:right;" title="`)
var topic_123 = []byte(`"></a>`)
var topic_124 = []byte(`

		</span>
	</div>
</article>

<div class="rowblock post_container" aria-label="`)
var topic_125 = []byte(`" style="overflow: hidden;">`)
var topic_126 = []byte(`
	<article itemscope itemtype="http://schema.org/CreativeWork" class="rowitem passive deletable_block editable_parent post_item action_item">
		<span class="action_icon" style="font-size: 18px;padding-right: 5px;">`)
var topic_127 = []byte(`</span>
		<span itemprop="text">`)
var topic_128 = []byte(`</span>
	</article>
`)
var topic_129 = []byte(`
	<article itemscope itemtype="http://schema.org/CreativeWork" class="rowitem passive deletable_block editable_parent post_item `)
var topic_130 = []byte(`" style="background-image: url(`)
var topic_131 = []byte(`), url(/static/`)
var topic_132 = []byte(`/post-avatar-bg.jpg);background-position: 0px `)
var topic_133 = []byte(`-1`)
var topic_134 = []byte(`0px;background-repeat:no-repeat, repeat-y;">
		`)
var topic_135 = []byte(`
		<p class="editable_block user_content" itemprop="text" style="margin:0;padding:0;">`)
var topic_136 = []byte(`</p>

		<span class="controls">

		<a href="`)
var topic_137 = []byte(`" class="username real_username" rel="author">`)
var topic_138 = []byte(`</a>&nbsp;&nbsp;
		`)
var topic_139 = []byte(`<a href="/reply/like/submit/`)
var topic_140 = []byte(`?session=`)
var topic_141 = []byte(`" class="mod_button" title="`)
var topic_142 = []byte(`" aria-label="`)
var topic_143 = []byte(`" style="color:#202020;"><button class="username like_label" style="background-color:#D6FFD6;"></button></a>`)
var topic_144 = []byte(`<a href="/reply/like/submit/`)
var topic_145 = []byte(`?session=`)
var topic_146 = []byte(`" class="mod_button" title="`)
var topic_147 = []byte(`" aria-label="`)
var topic_148 = []byte(`" style="color:#202020;"><button class="username like_label"></button></a>`)
var topic_149 = []byte(`<a href="/reply/edit/submit/`)
var topic_150 = []byte(`?session=`)
var topic_151 = []byte(`" class="mod_button" title="`)
var topic_152 = []byte(`" aria-label="`)
var topic_153 = []byte(`"><button class="username edit_item edit_label"></button></a>`)
var topic_154 = []byte(`<a href="/reply/delete/submit/`)
var topic_155 = []byte(`?session=`)
var topic_156 = []byte(`" class="mod_button" title="`)
var topic_157 = []byte(`" aria-label="`)
var topic_158 = []byte(`"><button class="username delete_item trash_label"></button></a>`)
var topic_159 = []byte(`<a class="mod_button" href='/users/ips/?ip=`)
var topic_160 = []byte(`' style="font-weight:normal;" title="`)
var topic_161 = []byte(`" aria-label="The poster's IP is `)
var topic_162 = []byte(`"><button class="username ip_label"></button></a>`)
var topic_163 = []byte(`
		<a href="/report/submit/`)
var topic_164 = []byte(`?session=`)
var topic_165 = []byte(`&type=reply" class="mod_button report_item" title="`)
var topic_166 = []byte(`" aria-label="`)
var topic_167 = []byte(`" rel="nofollow"><button class="username report_item flag_label"></button></a>

		`)
var topic_168 = []byte(`<a class="username hide_on_micro like_count">`)
var topic_169 = []byte(`</a><a class="username hide_on_micro like_count_label" title="`)
var topic_170 = []byte(`"></a>`)
var topic_171 = []byte(`<a class="username hide_on_micro user_tag">`)
var topic_172 = []byte(`</a>`)
var topic_173 = []byte(`<a class="username hide_on_micro level" aria-label="`)
var topic_174 = []byte(`">`)
var topic_175 = []byte(`</a><a class="username hide_on_micro level_label" style="float:right;" title="`)
var topic_176 = []byte(`"></a>`)
var topic_177 = []byte(`

		</span>
	</article>
`)
var topic_178 = []byte(`</div>

`)
var topic_179 = []byte(`
<div class="rowblock topic_reply_form quick_create_form" aria-label="`)
var topic_180 = []byte(`">
	<form id="quick_post_form" enctype="multipart/form-data" action="/reply/create/?session=`)
var topic_181 = []byte(`" method="post"></form>
	<input form="quick_post_form" name="tid" value='`)
var topic_182 = []byte(`' type="hidden" />
	<input form="quick_post_form" id="has_poll_input" name="has_poll" value="0" type="hidden" />
	<div class="formrow real_first_child">
		<div class="formitem">
			<textarea id="input_content" form="quick_post_form" name="reply-content" placeholder="`)
var topic_183 = []byte(`" required></textarea>
		</div>
	</div>
	<div class="formrow poll_content_row auto_hide">
		<div class="formitem">
			<div class="pollinput" data-pollinput="0">
				<input type="checkbox" disabled />
				<label class="pollinputlabel"></label>
				<input form="quick_post_form" name="pollinputitem[0]" class="pollinputinput" type="text" placeholder="`)
var topic_184 = []byte(`" />
			</div>
		</div>
	</div>
	<div class="formrow quick_button_row">
		<div class="formitem">
			<button form="quick_post_form" name="reply-button" class="formbutton">`)
var topic_185 = []byte(`</button>
			<button form="quick_post_form" class="formbutton" id="add_poll_button">`)
var topic_186 = []byte(`</button>
			`)
var topic_187 = []byte(`
			<input name="upload_files" form="quick_post_form" id="upload_files" multiple type="file" style="display: none;" />
			<label for="upload_files" class="formbutton add_file_button">`)
var topic_188 = []byte(`</label>
			<div id="upload_file_dock"></div>`)
var topic_189 = []byte(`
		</div>
	</div>
</div>
`)
var topic_190 = []byte(`

</main>

`)
var footer_0 = []byte(`<div class="footer">
	`)
var footer_1 = []byte(`
	<div id="poweredByHolder" class="footerBit">
		<div id="poweredBy">
			<a id="poweredByName" href="https://github.com/Azareal/Gosora">`)
var footer_2 = []byte(`</a><span id="poweredByDash"> - </span><span id="poweredByMaker">`)
var footer_3 = []byte(`</span>
		</div>
		<form action="/theme/" method="post">
			<div id="themeSelector" style="float: right;">
				<select id="themeSelectorSelect" name="themeSelector" aria-label="`)
var footer_4 = []byte(`">
				`)
var footer_5 = []byte(`<option val="`)
var footer_6 = []byte(`"`)
var footer_7 = []byte(` selected`)
var footer_8 = []byte(`>`)
var footer_9 = []byte(`</option>`)
var footer_10 = []byte(`
				</select>
			</div>
		</form>
	</div>
</div>
					</div>
				<aside class="sidebar">`)
var footer_11 = []byte(`</aside>
				<div style="clear: both;"></div>
			</div>
		</div>
	</body>
</html>
`)
var topic_alt_0 = []byte(`<link rel="prev" href="/topic/`)
var topic_alt_1 = []byte(`?page=`)
var topic_alt_2 = []byte(`" />
<div id="prevFloat" class="prev_button"><a class="prev_link" aria-label="`)
var topic_alt_3 = []byte(`" rel="prev" href="/topic/`)
var topic_alt_4 = []byte(`?page=`)
var topic_alt_5 = []byte(`">`)
var topic_alt_6 = []byte(`</a></div>`)
var topic_alt_7 = []byte(`<link rel="prerender next" href="/topic/`)
var topic_alt_8 = []byte(`?page=`)
var topic_alt_9 = []byte(`" />
<div id="nextFloat" class="next_button"><a class="next_link" aria-label="`)
var topic_alt_10 = []byte(`" rel="next" href="/topic/`)
var topic_alt_11 = []byte(`?page=`)
var topic_alt_12 = []byte(`">`)
var topic_alt_13 = []byte(`</a></div>`)
var topic_alt_14 = []byte(`

<main>

<div class="rowblock rowhead topic_block" aria-label="`)
var topic_alt_15 = []byte(`">
	<form action='/topic/edit/submit/`)
var topic_alt_16 = []byte(`?session=`)
var topic_alt_17 = []byte(`' method="post">
		<div class="rowitem topic_item`)
var topic_alt_18 = []byte(` topic_sticky_head`)
var topic_alt_19 = []byte(` topic_closed_head`)
var topic_alt_20 = []byte(`">
			<h1 class='topic_name hide_on_edit'>`)
var topic_alt_21 = []byte(`</h1>
			`)
var topic_alt_22 = []byte(`<span class='username hide_on_micro topic_status_e topic_status_closed hide_on_edit' title='`)
var topic_alt_23 = []byte(`' aria-label='`)
var topic_alt_24 = []byte(`' style="font-weight:normal;float: right;position:relative;top:-5px;">&#x1F512;&#xFE0E</span>`)
var topic_alt_25 = []byte(`
			<input class='show_on_edit topic_name_input' name="topic_name" value='`)
var topic_alt_26 = []byte(`' type="text" aria-label="`)
var topic_alt_27 = []byte(`" />
			<button name="topic-button" class="formbutton show_on_edit submit_edit">`)
var topic_alt_28 = []byte(`</button>
			`)
var topic_alt_29 = []byte(`
		</div>
	</form>
</div>

<div class="rowblock post_container">
	`)
var topic_alt_30 = []byte(`
	<form id="poll_`)
var topic_alt_31 = []byte(`_form" action="/poll/vote/`)
var topic_alt_32 = []byte(`?session=`)
var topic_alt_33 = []byte(`" method="post"></form>
	<article class="rowitem passive deletable_block editable_parent post_item poll_item top_post hide_on_edit">
		<div class="userinfo" aria-label="`)
var topic_alt_34 = []byte(`">
			<div class="avatar_item" style="background-image: url(`)
var topic_alt_35 = []byte(`), url(/static/white-dot.jpg);background-position: 0px -10px;">&nbsp;</div>
			<a href="`)
var topic_alt_36 = []byte(`" class="the_name" rel="author">`)
var topic_alt_37 = []byte(`</a>
			`)
var topic_alt_38 = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag">`)
var topic_alt_39 = []byte(`</div><div class="tag_post"></div></div>`)
var topic_alt_40 = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag post_level">`)
var topic_alt_41 = []byte(`</div><div class="tag_post"></div></div>`)
var topic_alt_42 = []byte(`
		</div>
		<div id="poll_voter_`)
var topic_alt_43 = []byte(`" class="content_container poll_voter">
			<div class="topic_content user_content">
				`)
var topic_alt_44 = []byte(`
				<div class="poll_option">
					<input form="poll_`)
var topic_alt_45 = []byte(`_form" id="poll_option_`)
var topic_alt_46 = []byte(`" name="poll_option_input" type="checkbox" value="`)
var topic_alt_47 = []byte(`" />
					<label class="poll_option_label" for="poll_option_`)
var topic_alt_48 = []byte(`">
						<div class="sel"></div>
					</label>
					<span id="poll_option_text_`)
var topic_alt_49 = []byte(`" class="poll_option_text">`)
var topic_alt_50 = []byte(`</span>
				</div>
				`)
var topic_alt_51 = []byte(`
				<div class="poll_buttons">
					<button form="poll_`)
var topic_alt_52 = []byte(`_form" class="poll_vote_button">`)
var topic_alt_53 = []byte(`</button>
					<button class="poll_results_button" data-poll-id="`)
var topic_alt_54 = []byte(`">`)
var topic_alt_55 = []byte(`</button>
					<a href="#"><button class="poll_cancel_button">`)
var topic_alt_56 = []byte(`</button></a>
				</div>
			</div>
		</div>
		<div id="poll_results_`)
var topic_alt_57 = []byte(`" class="content_container poll_results auto_hide">
			<div class="topic_content user_content"></div>
		</div>
	</article>
	`)
var topic_alt_58 = []byte(`
	<article itemscope itemtype="http://schema.org/CreativeWork" class="rowitem passive deletable_block editable_parent post_item top_post" aria-label="`)
var topic_alt_59 = []byte(`">
		<div class="userinfo" aria-label="`)
var topic_alt_60 = []byte(`">
			<div class="avatar_item" style="background-image: url(`)
var topic_alt_61 = []byte(`), url(/static/white-dot.jpg);background-position: 0px -10px;">&nbsp;</div>
			<a href="`)
var topic_alt_62 = []byte(`" class="the_name" rel="author">`)
var topic_alt_63 = []byte(`</a>
			`)
var topic_alt_64 = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag">`)
var topic_alt_65 = []byte(`</div><div class="tag_post"></div></div>`)
var topic_alt_66 = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag post_level">`)
var topic_alt_67 = []byte(`</div><div class="tag_post"></div></div>`)
var topic_alt_68 = []byte(`
		</div>
		<div class="content_container">
			<div class="hide_on_edit topic_content user_content" itemprop="text">`)
var topic_alt_69 = []byte(`</div>
			<textarea name="topic_content" class="show_on_edit topic_content_input">`)
var topic_alt_70 = []byte(`</textarea>
			<div class="button_container">
				`)
var topic_alt_71 = []byte(`<a href="/topic/like/submit/`)
var topic_alt_72 = []byte(`?session=`)
var topic_alt_73 = []byte(`" class="action_button like_item add_like" aria-label="`)
var topic_alt_74 = []byte(`" data-action="like"></a>`)
var topic_alt_75 = []byte(`<a href="/topic/edit/`)
var topic_alt_76 = []byte(`" class="action_button open_edit" aria-label="`)
var topic_alt_77 = []byte(`" data-action="edit"></a>`)
var topic_alt_78 = []byte(`<a href="/topic/delete/submit/`)
var topic_alt_79 = []byte(`?session=`)
var topic_alt_80 = []byte(`" class="action_button delete_item" aria-label="`)
var topic_alt_81 = []byte(`" data-action="delete"></a>`)
var topic_alt_82 = []byte(`<a href='/topic/unlock/submit/`)
var topic_alt_83 = []byte(`?session=`)
var topic_alt_84 = []byte(`' class="action_button unlock_item" data-action="unlock" aria-label="`)
var topic_alt_85 = []byte(`"></a>`)
var topic_alt_86 = []byte(`<a href='/topic/lock/submit/`)
var topic_alt_87 = []byte(`?session=`)
var topic_alt_88 = []byte(`' class="action_button lock_item" data-action="lock" aria-label="`)
var topic_alt_89 = []byte(`"></a>`)
var topic_alt_90 = []byte(`<a href='/topic/unstick/submit/`)
var topic_alt_91 = []byte(`?session=`)
var topic_alt_92 = []byte(`' class="action_button unpin_item" data-action="unpin" aria-label="`)
var topic_alt_93 = []byte(`"></a>`)
var topic_alt_94 = []byte(`<a href='/topic/stick/submit/`)
var topic_alt_95 = []byte(`?session=`)
var topic_alt_96 = []byte(`' class="action_button pin_item" data-action="pin" aria-label="`)
var topic_alt_97 = []byte(`"></a>`)
var topic_alt_98 = []byte(`<a href="/users/ips/?ip=`)
var topic_alt_99 = []byte(`" title="`)
var topic_alt_100 = []byte(`" class="action_button ip_item_button hide_on_big" aria-label="`)
var topic_alt_101 = []byte(`" data-action="ip"></a>`)
var topic_alt_102 = []byte(`
					<a href="/report/submit/`)
var topic_alt_103 = []byte(`?session=`)
var topic_alt_104 = []byte(`&type=topic" class="action_button report_item" aria-label="`)
var topic_alt_105 = []byte(`" data-action="report"></a>
					<a href="#" class="action_button button_menu"></a>
				`)
var topic_alt_106 = []byte(`
				<div class="action_button_right`)
var topic_alt_107 = []byte(` has_likes`)
var topic_alt_108 = []byte(`">
					`)
var topic_alt_109 = []byte(`<a class="action_button like_count hide_on_micro" aria-label="`)
var topic_alt_110 = []byte(`">`)
var topic_alt_111 = []byte(`</a>`)
var topic_alt_112 = []byte(`
					<a class="action_button created_at hide_on_mobile">`)
var topic_alt_113 = []byte(`</a>
					`)
var topic_alt_114 = []byte(`<a href="/users/ips/?ip=`)
var topic_alt_115 = []byte(`" title="`)
var topic_alt_116 = []byte(`" class="action_button ip_item hide_on_mobile" aria-hidden="true">`)
var topic_alt_117 = []byte(`</a>`)
var topic_alt_118 = []byte(`
				</div>
			</div>
		</div><div style="clear:both;"></div>
	</article>

	`)
var topic_alt_119 = []byte(`
	<article itemscope itemtype="http://schema.org/CreativeWork" class="rowitem passive deletable_block editable_parent post_item `)
var topic_alt_120 = []byte(`action_item`)
var topic_alt_121 = []byte(`">
		<div class="userinfo" aria-label="`)
var topic_alt_122 = []byte(`">
			<div class="avatar_item" style="background-image: url(`)
var topic_alt_123 = []byte(`), url(/static/white-dot.jpg);background-position: 0px -10px;">&nbsp;</div>
			<a href="`)
var topic_alt_124 = []byte(`" class="the_name" rel="author">`)
var topic_alt_125 = []byte(`</a>
			`)
var topic_alt_126 = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag">`)
var topic_alt_127 = []byte(`</div><div class="tag_post"></div></div>`)
var topic_alt_128 = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag post_level">`)
var topic_alt_129 = []byte(`</div><div class="tag_post"></div></div>`)
var topic_alt_130 = []byte(`
		</div>
		<div class="content_container" `)
var topic_alt_131 = []byte(`style="margin-left: 0px;"`)
var topic_alt_132 = []byte(`>
			`)
var topic_alt_133 = []byte(`
				<span class="action_icon" style="font-size: 18px;padding-right: 5px;" aria-hidden="true">`)
var topic_alt_134 = []byte(`</span>
				<span itemprop="text">`)
var topic_alt_135 = []byte(`</span>
			`)
var topic_alt_136 = []byte(`
			<div class="editable_block user_content" itemprop="text">`)
var topic_alt_137 = []byte(`</div>
			<div class="button_container">
				`)
var topic_alt_138 = []byte(`<a href="/reply/like/submit/`)
var topic_alt_139 = []byte(`?session=`)
var topic_alt_140 = []byte(`" class="action_button like_item add_like" aria-label="`)
var topic_alt_141 = []byte(`" data-action="like"></a>`)
var topic_alt_142 = []byte(`<a href="/reply/edit/submit/`)
var topic_alt_143 = []byte(`?session=`)
var topic_alt_144 = []byte(`" class="action_button edit_item" aria-label="`)
var topic_alt_145 = []byte(`" data-action="edit"></a>`)
var topic_alt_146 = []byte(`<a href="/reply/delete/submit/`)
var topic_alt_147 = []byte(`?session=`)
var topic_alt_148 = []byte(`" class="action_button delete_item" aria-label="`)
var topic_alt_149 = []byte(`" data-action="delete"></a>`)
var topic_alt_150 = []byte(`<a href="/users/ips/?ip=`)
var topic_alt_151 = []byte(`" title="`)
var topic_alt_152 = []byte(`" class="action_button ip_item_button hide_on_big" aria-label="`)
var topic_alt_153 = []byte(`" data-action="ip"></a>`)
var topic_alt_154 = []byte(`
					<a href="/report/submit/`)
var topic_alt_155 = []byte(`?session=`)
var topic_alt_156 = []byte(`&type=reply" class="action_button report_item" aria-label="`)
var topic_alt_157 = []byte(`" data-action="report"></a>
					<a href="#" class="action_button button_menu"></a>
				`)
var topic_alt_158 = []byte(`
				<div class="action_button_right`)
var topic_alt_159 = []byte(` has_likes`)
var topic_alt_160 = []byte(`">
					`)
var topic_alt_161 = []byte(`<a class="action_button like_count hide_on_micro" aria-label="`)
var topic_alt_162 = []byte(`">`)
var topic_alt_163 = []byte(`</a>`)
var topic_alt_164 = []byte(`
					<a class="action_button created_at hide_on_mobile">`)
var topic_alt_165 = []byte(`</a>
					`)
var topic_alt_166 = []byte(`<a href="/users/ips/?ip=`)
var topic_alt_167 = []byte(`" title="IP Address" class="action_button ip_item hide_on_mobile" aria-hidden="true">`)
var topic_alt_168 = []byte(`</a>`)
var topic_alt_169 = []byte(`
				</div>
			</div>
			`)
var topic_alt_170 = []byte(`
		</div>
		<div style="clear:both;"></div>
	</article>
`)
var topic_alt_171 = []byte(`</div>

`)
var topic_alt_172 = []byte(`
<div class="rowblock topic_reply_container">
	<div class="userinfo" aria-label="`)
var topic_alt_173 = []byte(`">
		<div class="avatar_item" style="background-image: url(`)
var topic_alt_174 = []byte(`), url(/static/white-dot.jpg);background-position: 0px -10px;">&nbsp;</div>
		<a href="`)
var topic_alt_175 = []byte(`" class="the_name" rel="author">`)
var topic_alt_176 = []byte(`</a>
		`)
var topic_alt_177 = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag">`)
var topic_alt_178 = []byte(`</div><div class="tag_post"></div></div>`)
var topic_alt_179 = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag post_level">`)
var topic_alt_180 = []byte(`</div><div class="tag_post"></div></div>`)
var topic_alt_181 = []byte(`
	</div>
	<div class="rowblock topic_reply_form quick_create_form"  aria-label="`)
var topic_alt_182 = []byte(`">
		<form id="quick_post_form" enctype="multipart/form-data" action="/reply/create/?session=`)
var topic_alt_183 = []byte(`" method="post"></form>
		<input form="quick_post_form" name="tid" value='`)
var topic_alt_184 = []byte(`' type="hidden" />
		<input form="quick_post_form" id="has_poll_input" name="has_poll" value="0" type="hidden" />
		<div class="formrow real_first_child">
			<div class="formitem">
				<textarea id="input_content" form="quick_post_form" name="reply-content" placeholder="`)
var topic_alt_185 = []byte(`" required></textarea>
			</div>
		</div>
		<div class="formrow poll_content_row auto_hide">
			<div class="formitem">
				<div class="pollinput" data-pollinput="0">
					<input type="checkbox" disabled />
					<label class="pollinputlabel"></label>
					<input form="quick_post_form" name="pollinputitem[0]" class="pollinputinput" type="text" placeholder="`)
var topic_alt_186 = []byte(`" />
				</div>
			</div>
		</div>
		<div class="formrow quick_button_row">
			<div class="formitem">
				<button form="quick_post_form" name="reply-button" class="formbutton">`)
var topic_alt_187 = []byte(`</button>
				<button form="quick_post_form" class="formbutton" id="add_poll_button">`)
var topic_alt_188 = []byte(`</button>
				`)
var topic_alt_189 = []byte(`
				<input name="upload_files" form="quick_post_form" id="upload_files" multiple type="file" style="display: none;" />
				<label for="upload_files" class="formbutton add_file_button">`)
var topic_alt_190 = []byte(`</label>
				<div id="upload_file_dock"></div>`)
var topic_alt_191 = []byte(`
			</div>
		</div>
	</div>
</div>
`)
var topic_alt_192 = []byte(`

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
			`)
var profile_8 = []byte(`<div class="rowitem passive">
				<a class="profile_menu_item">`)
var profile_9 = []byte(`</a>
			</div>`)
var profile_10 = []byte(`
			<!--<div class="rowitem passive">
				<a class="profile_menu_item">`)
var profile_11 = []byte(`</a>
			</div>-->
			`)
var profile_12 = []byte(`<div class="rowitem passive">
				`)
var profile_13 = []byte(`<a href="/users/unban/`)
var profile_14 = []byte(`?session=`)
var profile_15 = []byte(`" class="profile_menu_item">`)
var profile_16 = []byte(`</a>
			`)
var profile_17 = []byte(`<a href="#ban_user" class="profile_menu_item">`)
var profile_18 = []byte(`</a>`)
var profile_19 = []byte(`
			</div>`)
var profile_20 = []byte(`
			<div class="rowitem passive">
				<a href="/report/submit/`)
var profile_21 = []byte(`?session=`)
var profile_22 = []byte(`&type=user" class="profile_menu_item report_item" aria-label="`)
var profile_23 = []byte(`" title="`)
var profile_24 = []byte(`"></a>
			</div>
			`)
var profile_25 = []byte(`
		</div>
	</div>
</div>

<div id="profile_right_lane" class="colstack_right">
	`)
var profile_26 = []byte(`
	<!-- TODO: Inline the display: none; CSS -->
	<div id="ban_user_head" class="colstack_item colstack_head hash_hide ban_user_hash" style="display: none;">
			<div class="rowitem"><h1><a>`)
var profile_27 = []byte(`</a></h1></div>
	</div>
	<form id="ban_user_form" class="hash_hide ban_user_hash" action="/users/ban/submit/`)
var profile_28 = []byte(`?session=`)
var profile_29 = []byte(`" method="post" style="display: none;">
		`)
var profile_30 = []byte(`
		<div class="colline">`)
var profile_31 = []byte(`</div>
		<div class="colstack_item">
			<div class="formrow real_first_child">
				<div class="formitem formlabel"><a>`)
var profile_32 = []byte(`</a></div>
				<div class="formitem">
					<input name="ban-duration-days" type="number" value="0" min="0" />
				</div>
			</div>
			<div class="formrow">
				<div class="formitem formlabel"><a>`)
var profile_33 = []byte(`</a></div>
				<div class="formitem">
					<input name="ban-duration-weeks" type="number" value="0" min="0" />
				</div>
			</div>
			<div class="formrow">
				<div class="formitem formlabel"><a>`)
var profile_34 = []byte(`</a></div>
				<div class="formitem">
					<input name="ban-duration-months" type="number" value="0" min="0" />
				</div>
			</div>
			<!--<div class="formrow">
				<div class="formitem formlabel"><a>`)
var profile_35 = []byte(`</a></div>
				<div class="formitem"><textarea name="ban-reason" placeholder="A really horrible person" required></textarea></div>
			</div>-->
			<div class="formrow">
				<div class="formitem"><button name="ban-button" class="formbutton form_middle_button">`)
var profile_36 = []byte(`</button></div>
			</div>
		</div>
	</form>
	`)
var profile_37 = []byte(`

	<div id="profile_comments_head" class="colstack_item colstack_head hash_hide">
		<div class="rowitem"><h1><a>`)
var profile_38 = []byte(`</a></h1></div>
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
var profile_comments_row_10 = []byte(`" class="mod_button" title="`)
var profile_comments_row_11 = []byte(`" aria-label="`)
var profile_comments_row_12 = []byte(`"><button class="username edit_item edit_label"></button></a>

				<a href="/profile/reply/delete/submit/`)
var profile_comments_row_13 = []byte(`?session=`)
var profile_comments_row_14 = []byte(`" class="mod_button" title="`)
var profile_comments_row_15 = []byte(`" aria-label="`)
var profile_comments_row_16 = []byte(`"><button class="username delete_item trash_label"></button></a>`)
var profile_comments_row_17 = []byte(`

				<a class="mod_button" href="/report/submit/`)
var profile_comments_row_18 = []byte(`?session=`)
var profile_comments_row_19 = []byte(`&type=user-reply"><button class="username report_item flag_label" title="`)
var profile_comments_row_20 = []byte(`" aria-label="`)
var profile_comments_row_21 = []byte(`"></button></a>

				`)
var profile_comments_row_22 = []byte(`<a class="username hide_on_mobile user_tag" style="float: right;">`)
var profile_comments_row_23 = []byte(`</a>`)
var profile_comments_row_24 = []byte(`
			</span>
		</div>
	`)
var profile_comments_row_25 = []byte(`
		<div class="rowitem passive deletable_block editable_parent comment `)
var profile_comments_row_26 = []byte(`">
			<div class="topRow">
				<div class="userbit">
					<img src="`)
var profile_comments_row_27 = []byte(`" alt="`)
var profile_comments_row_28 = []byte(`'s Avatar" title="`)
var profile_comments_row_29 = []byte(`'s Avatar" />
					<span class="nameAndTitle">
						<a href="`)
var profile_comments_row_30 = []byte(`" class="real_username username">`)
var profile_comments_row_31 = []byte(`</a>
						`)
var profile_comments_row_32 = []byte(`<a class="username hide_on_mobile user_tag" style="float: right;">`)
var profile_comments_row_33 = []byte(`</a>`)
var profile_comments_row_34 = []byte(`
					</span>
				</div>
				<span class="controls">
					`)
var profile_comments_row_35 = []byte(`
						<a href="/profile/reply/edit/submit/`)
var profile_comments_row_36 = []byte(`?session=`)
var profile_comments_row_37 = []byte(`" class="mod_button" title="`)
var profile_comments_row_38 = []byte(`" aria-label="`)
var profile_comments_row_39 = []byte(`"><button class="username edit_item edit_label"></button></a>
						<a href="/profile/reply/delete/submit/`)
var profile_comments_row_40 = []byte(`?session=`)
var profile_comments_row_41 = []byte(`" class="mod_button" title="`)
var profile_comments_row_42 = []byte(`" aria-label="`)
var profile_comments_row_43 = []byte(`"><button class="username delete_item trash_label"></button></a>
					`)
var profile_comments_row_44 = []byte(`
					<a class="mod_button" href="/report/submit/`)
var profile_comments_row_45 = []byte(`?session=`)
var profile_comments_row_46 = []byte(`&type=user-reply"><button class="username report_item flag_label" title="`)
var profile_comments_row_47 = []byte(`" aria-label="`)
var profile_comments_row_48 = []byte(`"></button></a>
				</span>
			</div>
			<div class="content_column">
				<span class="editable_block user_content">`)
var profile_comments_row_49 = []byte(`</span>
			</div>
		</div>
		<div class="after_comment"></div>
	`)
var profile_39 = []byte(`</div>

`)
var profile_40 = []byte(`
	<form id="profile_comments_form" class="hash_hide" action="/profile/reply/create/?session=`)
var profile_41 = []byte(`" method="post">
		<input name="uid" value='`)
var profile_42 = []byte(`' type="hidden" />
		<div class="colstack_item topic_reply_form" style="border-top: none;">
			<div class="formrow">
				<div class="formitem"><textarea class="input_content" name="reply-content" placeholder="`)
var profile_43 = []byte(`"></textarea></div>
			</div>
			<div class="formrow quick_button_row">
				<div class="formitem"><button name="reply-button" class="formbutton">`)
var profile_44 = []byte(`</button></div>
			</div>
		</div>
	</form>
`)
var profile_45 = []byte(`
</div>

</div>

`)
var profile_46 = []byte(`
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
<main id="forumsItemList" itemscope itemtype="http://schema.org/ItemList">

<div class="rowblock opthead">
	<div class="rowitem"><h1 itemprop="name">`)
var forums_1 = []byte(`</h1></div>
</div>
<div class="rowblock forum_list">
	`)
var forums_2 = []byte(`<div class="rowitem `)
var forums_3 = []byte(`datarow `)
var forums_4 = []byte(`"itemprop="itemListElement" itemscope
      itemtype="http://schema.org/ListItem">
		<span class="forum_left shift_left">
			<a href="`)
var forums_5 = []byte(`" itemprop="item">`)
var forums_6 = []byte(`</a>
		`)
var forums_7 = []byte(`
			<br /><span class="rowsmall" itemprop="description">`)
var forums_8 = []byte(`</span>
		`)
var forums_9 = []byte(`
			<br /><span class="rowsmall" style="font-style: italic;">`)
var forums_10 = []byte(`</span>
		`)
var forums_11 = []byte(`
		</span>

		<span class="forum_right shift_right">
			`)
var forums_12 = []byte(`<img class="extra_little_row_avatar" src="`)
var forums_13 = []byte(`" height=64 width=64 alt="`)
var forums_14 = []byte(`'s Avatar" title="`)
var forums_15 = []byte(`'s Avatar" />`)
var forums_16 = []byte(`
			<span>
				<a href="`)
var forums_17 = []byte(`">`)
var forums_18 = []byte(`</a>
				`)
var forums_19 = []byte(`<br /><span class="rowsmall">`)
var forums_20 = []byte(`</span>`)
var forums_21 = []byte(`
			</span>
		</span>
		<div style="clear: both;"></div>
	</div>
	`)
var forums_22 = []byte(`<div class="rowitem passive rowmsg">`)
var forums_23 = []byte(`</div>`)
var forums_24 = []byte(`
</div>

</main>
`)
var topics_0 = []byte(`
<main id="topicsItemList" itemscope itemtype="http://schema.org/ItemList">

<div class="rowblock rowhead topic_list_title_block`)
var topics_1 = []byte(` has_opt`)
var topics_2 = []byte(`">
	<div class="rowitem topic_list_title"><h1 itemprop="name">`)
var topics_3 = []byte(`</h1></div>
	`)
var topics_4 = []byte(`
		<div class="optbox">
		`)
var topics_5 = []byte(`
			<div class="pre_opt auto_hide"></div>
			<div class="opt create_topic_opt" title="`)
var topics_6 = []byte(`" aria-label="`)
var topics_7 = []byte(`"><a class="create_topic_link" href="/topics/create/"></a></div>
			`)
var topics_8 = []byte(`
			<div class="opt mod_opt" title="`)
var topics_9 = []byte(`">
				<a class="moderate_link" href="#" aria-label="`)
var topics_10 = []byte(`"></a>
			</div>
			`)
var topics_11 = []byte(`<div class="opt locked_opt" title="`)
var topics_12 = []byte(`" aria-label="`)
var topics_13 = []byte(`"><a></a></div>`)
var topics_14 = []byte(`
		</div>
		<div style="clear: both;"></div>
	`)
var topics_15 = []byte(`
</div>

`)
var topics_16 = []byte(`
<div class="mod_floater auto_hide">
	<form method="post">
	<div class="mod_floater_head">
		<span>`)
var topics_17 = []byte(`</span>
	</div>
	<div class="mod_floater_body">
		<select class="mod_floater_options">
			<option val="delete">`)
var topics_18 = []byte(`</option>
			<option val="lock">`)
var topics_19 = []byte(`</option>
			<option val="move">`)
var topics_20 = []byte(`</option>
		</select>
		<button class="mod_floater_submit">`)
var topics_21 = []byte(`</button>
	</div>
	</form>
</div>

`)
var topics_22 = []byte(`
<div id="mod_topic_mover" class="modal_pane auto_hide">
	<form action="/topic/move/submit/?session=`)
var topics_23 = []byte(`" method="post">
		<input id="mover_fid" name="fid" value="0" type="hidden" />
		<div class="pane_header">
			<h3>`)
var topics_24 = []byte(`</h3>
		</div>
		<div class="pane_body">
			<div class="pane_table">
				`)
var topics_25 = []byte(`<div id="mover_fid_`)
var topics_26 = []byte(`" data-fid="`)
var topics_27 = []byte(`" class="pane_row">`)
var topics_28 = []byte(`</div>`)
var topics_29 = []byte(`
			</div>
		</div>
		<div class="pane_buttons">
			<button id="mover_submit">`)
var topics_30 = []byte(`</button>
		</div>
	</form>
</div>
<div class="rowblock topic_create_form quick_create_form" style="display: none;" aria-label="`)
var topics_31 = []byte(`">
	<form name="topic_create_form_form" id="quick_post_form" enctype="multipart/form-data" action="/topic/create/submit/?session=`)
var topics_32 = []byte(`" method="post"></form>
	<input form="quick_post_form" id="has_poll_input" name="has_poll" value="0" type="hidden" />
	<img class="little_row_avatar" src="`)
var topics_33 = []byte(`" height="64" alt="`)
var topics_34 = []byte(`" title="`)
var topics_35 = []byte(`" />
	<div class="main_form">
		<div class="topic_meta">
			<div class="formrow topic_board_row real_first_child">
				<div class="formitem"><select form="quick_post_form" id="topic_board_input" name="topic-board">
					`)
var topics_36 = []byte(`<option `)
var topics_37 = []byte(`selected`)
var topics_38 = []byte(` value="`)
var topics_39 = []byte(`">`)
var topics_40 = []byte(`</option>`)
var topics_41 = []byte(`
				</select></div>
			</div>
			<div class="formrow topic_name_row">
				<div class="formitem">
					<input form="quick_post_form" name="topic-name" placeholder="`)
var topics_42 = []byte(`" required>
				</div>
			</div>
		</div>
		<div class="formrow topic_content_row">
			<div class="formitem">
				<textarea form="quick_post_form" id="input_content" name="topic-content" placeholder="`)
var topics_43 = []byte(`" required></textarea>
			</div>
		</div>
		<div class="formrow poll_content_row auto_hide">
			<div class="formitem">
				<div class="pollinput" data-pollinput="0">
					<input type="checkbox" disabled />
					<label class="pollinputlabel"></label>
					<input form="quick_post_form" name="pollinputitem[0]" class="pollinputinput" type="text" placeholder="`)
var topics_44 = []byte(`" />
				</div>
			</div>
		</div>
		<div class="formrow quick_button_row">
			<div class="formitem">
				<button form="quick_post_form" class="formbutton">`)
var topics_45 = []byte(`</button>
				<button form="quick_post_form" class="formbutton" id="add_poll_button">`)
var topics_46 = []byte(`</button>
				`)
var topics_47 = []byte(`
				<input name="upload_files" form="quick_post_form" id="upload_files" multiple type="file" style="display: none;" />
				<label for="upload_files" class="formbutton add_file_button">`)
var topics_48 = []byte(`</label>
				<div id="upload_file_dock"></div>`)
var topics_49 = []byte(`
				<button class="formbutton close_form">`)
var topics_50 = []byte(`</button>
			</div>
		</div>
	</div>
</div>
	`)
var topics_51 = []byte(`
<div id="topic_list" class="rowblock topic_list" aria-label="`)
var topics_52 = []byte(`">
	`)
var topics_53 = []byte(`<div class="topic_row" data-tid="`)
var topics_54 = []byte(`">
	<div class="rowitem topic_left passive datarow `)
var topics_55 = []byte(`topic_sticky`)
var topics_56 = []byte(`topic_closed`)
var topics_57 = []byte(`">
		<span class="selector"></span>
		<a href="`)
var topics_58 = []byte(`"><img src="`)
var topics_59 = []byte(`" height="64" alt="`)
var topics_60 = []byte(`'s Avatar" title="`)
var topics_61 = []byte(`'s Avatar" /></a>
		<span class="topic_inner_left">
			<a class="rowtopic" href="`)
var topics_62 = []byte(`" itemprop="itemListElement"><span>`)
var topics_63 = []byte(`</span></a> `)
var topics_64 = []byte(`<a class="rowsmall parent_forum" href="`)
var topics_65 = []byte(`">`)
var topics_66 = []byte(`</a>`)
var topics_67 = []byte(`
			<br /><a class="rowsmall starter" href="`)
var topics_68 = []byte(`">`)
var topics_69 = []byte(`</a>
			`)
var topics_70 = []byte(`<span class="rowsmall topic_status_e topic_status_closed" title="`)
var topics_71 = []byte(`"> | &#x1F512;&#xFE0E</span>`)
var topics_72 = []byte(`<span class="rowsmall topic_status_e topic_status_sticky" title="`)
var topics_73 = []byte(`"> | &#x1F4CD;&#xFE0E</span>`)
var topics_74 = []byte(`
		</span>
		<span class="topic_inner_right rowsmall" style="float: right;">
			<span class="replyCount">`)
var topics_75 = []byte(`</span><br />
			<span class="likeCount">`)
var topics_76 = []byte(`</span>
		</span>
	</div>
	<div class="rowitem topic_right passive datarow `)
var topics_77 = []byte(`topic_sticky`)
var topics_78 = []byte(`topic_closed`)
var topics_79 = []byte(`">
		<a href="`)
var topics_80 = []byte(`"><img src="`)
var topics_81 = []byte(`" height="64" alt="`)
var topics_82 = []byte(`'s Avatar" title="`)
var topics_83 = []byte(`'s Avatar" /></a>
		<span>
			<a href="`)
var topics_84 = []byte(`" class="lastName" style="font-size: 14px;">`)
var topics_85 = []byte(`</a><br>
			<span class="rowsmall lastReplyAt">`)
var topics_86 = []byte(`</span>
		</span>
	</div>
	</div>`)
var topics_87 = []byte(`<div class="rowitem passive rowmsg">`)
var topics_88 = []byte(` <a href="/topics/create/">`)
var topics_89 = []byte(`</a>`)
var topics_90 = []byte(`</div>`)
var topics_91 = []byte(`
</div>

`)
var paginator_0 = []byte(`<div class="pageset">
	`)
var paginator_1 = []byte(`<div class="pageitem"><a href="?page=`)
var paginator_2 = []byte(`" rel="prev" aria-label="`)
var paginator_3 = []byte(`">`)
var paginator_4 = []byte(`</a></div>
	<link rel="prev" href="?page=`)
var paginator_5 = []byte(`" />`)
var paginator_6 = []byte(`
	<div class="pageitem"><a href="?page=`)
var paginator_7 = []byte(`">`)
var paginator_8 = []byte(`</a></div>
	`)
var paginator_9 = []byte(`
	<link rel="next" href="?page=`)
var paginator_10 = []byte(`" />
	<div class="pageitem"><a href="?page=`)
var paginator_11 = []byte(`" rel="next" aria-label="`)
var paginator_12 = []byte(`">`)
var paginator_13 = []byte(`</a></div>`)
var paginator_14 = []byte(`
</div>`)
var topics_92 = []byte(`

</main>
`)
var forum_0 = []byte(`<div id="prevFloat" class="prev_button"><a class="prev_link" aria-label="`)
var forum_1 = []byte(`" rel="prev" href="/forum/`)
var forum_2 = []byte(`?page=`)
var forum_3 = []byte(`">`)
var forum_4 = []byte(`</a></div>`)
var forum_5 = []byte(`<div id="nextFloat" class="next_button"><a class="next_link" aria-label="`)
var forum_6 = []byte(`" rel="next" href="/forum/`)
var forum_7 = []byte(`?page=`)
var forum_8 = []byte(`">`)
var forum_9 = []byte(`</a></div>`)
var forum_10 = []byte(`

<main id="forumItemList" itemscope itemtype="http://schema.org/ItemList">
	<div id="forum_head_block" class="rowblock rowhead topic_list_title_block`)
var forum_11 = []byte(` has_opt`)
var forum_12 = []byte(`">
		<div class="rowitem forum_title">
			<h1 itemprop="name">`)
var forum_13 = []byte(`</h1>
		</div>
		`)
var forum_14 = []byte(`
			<div class="optbox">
				`)
var forum_15 = []byte(`
				<div class="pre_opt auto_hide"></div>
				<div class="opt create_topic_opt" title="`)
var forum_16 = []byte(`" aria-label="`)
var forum_17 = []byte(`"><a class="create_topic_link" href="/topics/create/`)
var forum_18 = []byte(`"></a></div>
				`)
var forum_19 = []byte(`
				<div class="opt mod_opt" title="`)
var forum_20 = []byte(`">
					<a class="moderate_link" href="#" aria-label="`)
var forum_21 = []byte(`"></a>
				</div>
				`)
var forum_22 = []byte(`<div class="opt locked_opt" title="`)
var forum_23 = []byte(`" aria-label="`)
var forum_24 = []byte(`"><a></a></div>`)
var forum_25 = []byte(`
			</div>
			<div style="clear: both;"></div>
		`)
var forum_26 = []byte(`
	</div>
	`)
var forum_27 = []byte(`
	<div class="mod_floater auto_hide">
		<form method="post">
			<div class="mod_floater_head">
				<span>`)
var forum_28 = []byte(`</span>
			</div>
			<div class="mod_floater_body">
				<select class="mod_floater_options">
					<option val="delete">`)
var forum_29 = []byte(`</option>
					<option val="lock">`)
var forum_30 = []byte(`</option>
					<option val="move">`)
var forum_31 = []byte(`</option>
				</select>
				<button>`)
var forum_32 = []byte(`</button>
			</div>
		</form>
	</div>
	`)
var forum_33 = []byte(`
	<div id="forum_topic_create_form" class="rowblock topic_create_form quick_create_form" style="display: none;" aria-label="`)
var forum_34 = []byte(`">
		<form id="quick_post_form" enctype="multipart/form-data" action="/topic/create/submit/" method="post"></form>
		<img class="little_row_avatar" src="`)
var forum_35 = []byte(`" height="64" alt="`)
var forum_36 = []byte(`" title="`)
var forum_37 = []byte(`" />
		<input form="quick_post_form" id="topic_board_input" name="topic-board" value="`)
var forum_38 = []byte(`" type="hidden">
		<div class="main_form">
			<div class="topic_meta">
				<div class="formrow topic_name_row real_first_child">
					<div class="formitem">
						<input form="quick_post_form" name="topic-name" placeholder="`)
var forum_39 = []byte(`" required>
					</div>
				</div>
			</div>
			<div class="formrow topic_content_row">
				<div class="formitem">
					<textarea form="quick_post_form" id="input_content" name="topic-content" placeholder="`)
var forum_40 = []byte(`" required></textarea>
				</div>
			</div>
			<div class="formrow poll_content_row auto_hide">
				<div class="formitem">
					Poll stuff
				</div>
			</div>
			<div class="formrow quick_button_row">
				<div class="formitem">
					<button form="quick_post_form" name="topic-button" class="formbutton">`)
var forum_41 = []byte(`</button>
					<button form="quick_post_form" class="formbutton" id="add_poll_button">`)
var forum_42 = []byte(`</button>
					`)
var forum_43 = []byte(`
					<input name="upload_files" form="quick_post_form" id="upload_files" multiple type="file" style="display: none;" />
					<label for="upload_files" class="formbutton add_file_button">`)
var forum_44 = []byte(`</label>
					<div id="upload_file_dock"></div>`)
var forum_45 = []byte(`
					<button class="formbutton close_form">`)
var forum_46 = []byte(`</button>
				</div>
			</div>
		</div>
	</div>
	`)
var forum_47 = []byte(`
	<div id="forum_topic_list" class="rowblock topic_list" aria-label="`)
var forum_48 = []byte(`">
		`)
var forum_49 = []byte(`<div class="topic_row" data-tid="`)
var forum_50 = []byte(`">
		<div class="rowitem topic_left passive datarow `)
var forum_51 = []byte(`topic_sticky`)
var forum_52 = []byte(`topic_closed`)
var forum_53 = []byte(`">
			<span class="selector"></span>
			<a href="`)
var forum_54 = []byte(`"><img src="`)
var forum_55 = []byte(`" height="64" alt="`)
var forum_56 = []byte(`'s Avatar" title="`)
var forum_57 = []byte(`'s Avatar" /></a>
			<span class="topic_inner_left">
				<a class="rowtopic" href="`)
var forum_58 = []byte(`" itemprop="itemListElement"><span>`)
var forum_59 = []byte(`</span></a>
				<br /><a class="rowsmall starter" href="`)
var forum_60 = []byte(`">`)
var forum_61 = []byte(`</a>
				`)
var forum_62 = []byte(`<span class="rowsmall topic_status_e topic_status_closed" title="`)
var forum_63 = []byte(`"> | &#x1F512;&#xFE0E</span>`)
var forum_64 = []byte(`<span class="rowsmall topic_status_e topic_status_sticky" title="`)
var forum_65 = []byte(`"> | &#x1F4CD;&#xFE0E</span>`)
var forum_66 = []byte(`
			</span>
			<span class="topic_inner_right rowsmall" style="float: right;">
				<span class="replyCount">`)
var forum_67 = []byte(`</span><br />
				<span class="likeCount">`)
var forum_68 = []byte(`</span>
			</span>
		</div>
		<div class="rowitem topic_right passive datarow `)
var forum_69 = []byte(`topic_sticky`)
var forum_70 = []byte(`topic_closed`)
var forum_71 = []byte(`">
			<a href="`)
var forum_72 = []byte(`"><img src="`)
var forum_73 = []byte(`" height="64" alt="`)
var forum_74 = []byte(`'s Avatar" title="`)
var forum_75 = []byte(`'s Avatar" /></a>
			<span>
				<a href="`)
var forum_76 = []byte(`" class="lastName" style="font-size: 14px;">`)
var forum_77 = []byte(`</a><br>
				<span class="rowsmall lastReplyAt">`)
var forum_78 = []byte(`</span>
			</span>
		</div>
		</div>`)
var forum_79 = []byte(`<div class="rowitem passive rowmsg">`)
var forum_80 = []byte(` <a href="/topics/create/`)
var forum_81 = []byte(`">`)
var forum_82 = []byte(`</a>`)
var forum_83 = []byte(`</div>`)
var forum_84 = []byte(`
	</div>

`)
var forum_85 = []byte(`

</main>
`)
var login_0 = []byte(`
<main id="login_page">
	<div class="rowblock rowhead">
		<div class="rowitem"><h1>`)
var login_1 = []byte(`</h1></div>
	</div>
	<div class="rowblock">
		<form action="/accounts/login/submit/" method="post">
			<div class="formrow login_name_row">
				<div class="formitem formlabel"><a id="login_name_label">`)
var login_2 = []byte(`</a></div>
				<div class="formitem"><input name="username" type="text" placeholder="`)
var login_3 = []byte(`" aria-labelledby="login_name_label" required /></div>
			</div>
			<div class="formrow login_password_row">
				<div class="formitem formlabel"><a id="login_password_label">`)
var login_4 = []byte(`</a></div>
				<div class="formitem"><input name="password" type="password" autocomplete="current-password" placeholder="*****" aria-labelledby="login_password_label" required /></div>
			</div>
			<div class="formrow login_button_row">
				<div class="formitem"><button name="login-button" class="formbutton">`)
var login_5 = []byte(`</button></div>
				<div class="formitem dont_have_account">`)
var login_6 = []byte(`</div>
			</div>
		</form>
	</div>
</main>
`)
var register_0 = []byte(`
<main id="register_page">
	<div class="rowblock rowhead">
		<div class="rowitem"><h1>`)
var register_1 = []byte(`</h1></div>
	</div>
	<div class="rowblock">
		<form action="/accounts/create/submit/" method="post">
			<div class="formrow">
				<div class="formitem formlabel"><a id="username_label">`)
var register_2 = []byte(`</a></div>
				<div class="formitem"><input name="username" type="text" placeholder="`)
var register_3 = []byte(`" aria-labelledby="username_label" required /></div>
			</div>
			<div class="formrow">
				<div class="formitem formlabel"><a id="email_label">`)
var register_4 = []byte(`</a></div>
				<div class="formitem"><input name="email" type="email" placeholder="joe.doe@example.com" aria-labelledby="email_label" required /></div>
			</div>
			<div class="formrow">
				<div class="formitem formlabel"><a id="password_label">`)
var register_5 = []byte(`</a></div>
				<div class="formitem"><input name="password" type="password" autocomplete="new-password" placeholder="*****" aria-labelledby="password_label" required /></div>
			</div>
			<div class="formrow">
				<div class="formitem formlabel"><a id="confirm_password_label">`)
var register_6 = []byte(`</a></div>
				<div class="formitem"><input name="confirm_password" type="password" placeholder="*****" aria-labelledby="confirm_password_label" required /></div>
			</div>
			<div class="formrow">
				<div class="formitem"><button name="register-button" class="formbutton">`)
var register_7 = []byte(`</button></div>
			</div>
		</form>
	</div>
</main>
`)
var error_0 = []byte(`
<main>
	<div class="rowblock rowhead">
		<div class="rowitem"><h1>`)
var error_1 = []byte(`</h1></div>
	</div>
	<div class="rowblock">
		<div class="rowitem passive rowmsg">`)
var error_2 = []byte(`</div>
	</div>
</main>
`)
var ip_search_0 = []byte(`
<main id="ip_search_container">
	<div class="rowblock rowhead">
		<div class="rowitem">
			<h1>`)
var ip_search_1 = []byte(`</h1>
		</div>
	</div>
	<form action="/users/ips/" method="get" id="ip-search-form"></form>
	<div class="rowblock ip_search_block">
		<div class="rowitem passive">
			<input form="ip-search-form" name="ip" class="ip_search_input" type="search" placeholder="🔍︎"`)
var ip_search_2 = []byte(` value="`)
var ip_search_3 = []byte(`"`)
var ip_search_4 = []byte(` />
			<input form="ip-search-form" class="ip_search_search" type="submit" value="`)
var ip_search_5 = []byte(`" />
		</div>
	</div>
	`)
var ip_search_6 = []byte(`
	<div class="rowblock rowlist bgavatars">
		`)
var ip_search_7 = []byte(`<div class="rowitem" style="background-image: url('`)
var ip_search_8 = []byte(`');">
			<img src="`)
var ip_search_9 = []byte(`" class="bgsub" alt="`)
var ip_search_10 = []byte(`'s Avatar" />
			<a class="rowTitle" href="`)
var ip_search_11 = []byte(`">`)
var ip_search_12 = []byte(`</a>
		</div>
		`)
var ip_search_13 = []byte(`<div class="rowitem rowmsg">`)
var ip_search_14 = []byte(`</div>`)
var ip_search_15 = []byte(`
	</div>
	`)
var ip_search_16 = []byte(`
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
