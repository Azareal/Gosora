package main

var error_frags = make([][]byte,4)

// nolint
func Get_error_frags() [][]byte {
return error_frags
}
var profile_comments_row_frags = make([][]byte,51)

// nolint
func Get_profile_comments_row_frags() [][]byte {
return profile_comments_row_frags
}
var topic_alt_frags = make([][]byte,200)

// nolint
func Get_topic_alt_frags() [][]byte {
return topic_alt_frags
}
var profile_frags = make([][]byte,50)

// nolint
func Get_profile_frags() [][]byte {
return profile_frags
}
var ip_search_frags = make([][]byte,18)

// nolint
func Get_ip_search_frags() [][]byte {
return ip_search_frags
}
var header_frags = make([][]byte,26)

// nolint
func Get_header_frags() [][]byte {
return header_frags
}
var forums_frags = make([][]byte,26)

// nolint
func Get_forums_frags() [][]byte {
return forums_frags
}
var login_frags = make([][]byte,8)

// nolint
func Get_login_frags() [][]byte {
return login_frags
}
var register_frags = make([][]byte,9)

// nolint
func Get_register_frags() [][]byte {
return register_frags
}
var footer_frags = make([][]byte,13)

// nolint
func Get_footer_frags() [][]byte {
return footer_frags
}
var topic_frags = make([][]byte,199)

// nolint
func Get_topic_frags() [][]byte {
return topic_frags
}
var paginator_frags = make([][]byte,16)

// nolint
func Get_paginator_frags() [][]byte {
return paginator_frags
}
var topics_frags = make([][]byte,98)

// nolint
func Get_topics_frags() [][]byte {
return topics_frags
}
var forum_frags = make([][]byte,90)

// nolint
func Get_forum_frags() [][]byte {
return forum_frags
}
var guilds_guild_list_frags = make([][]byte,10)

// nolint
func Get_guilds_guild_list_frags() [][]byte {
return guilds_guild_list_frags
}
var notice_frags = make([][]byte,3)

// nolint
func Get_notice_frags() [][]byte {
return notice_frags
}

// nolint
func init() {
header_frags[0] = []byte(`<!doctype html>
<html lang="en">
	<head>
		<title>`)
header_frags[1] = []byte(` | `)
header_frags[2] = []byte(`</title>
		<link href="/static/`)
header_frags[3] = []byte(`/main.css" rel="stylesheet" type="text/css">
		`)
header_frags[4] = []byte(`
		<link href="/static/`)
header_frags[5] = []byte(`" rel="stylesheet" type="text/css">
		`)
header_frags[6] = []byte(`
		<script type="text/javascript" src="/static/jquery-3.1.1.min.js"></script>
		<script type="text/javascript" src="/static/chartist/chartist.min.js"></script>
		`)
header_frags[7] = []byte(`
		<script type="text/javascript" src="/static/`)
header_frags[8] = []byte(`"></script>
		`)
header_frags[9] = []byte(`
		<script type="text/javascript">
		var session = "`)
header_frags[10] = []byte(`";
		var siteURL = "`)
header_frags[11] = []byte(`";
		</script>
		<script type="text/javascript" src="/static/global.js"></script>
		<meta name="viewport" content="width=device-width,initial-scale = 1.0, maximum-scale=1.0,user-scalable=no" />
		`)
header_frags[12] = []byte(`<meta name="description" content="`)
header_frags[13] = []byte(`" />`)
header_frags[14] = []byte(`
	</head>
	<body>
		`)
header_frags[15] = []byte(`<style>.supermod_only { display: none !important; }</style>`)
header_frags[16] = []byte(`
		<div class="container">
			<div class="left_of_nav">`)
header_frags[17] = []byte(`</div>
			<nav class="nav">
				<div class="move_left">
				<div class="move_right">
				<ul>
					<li id="menu_overview" class="menu_left"><a href="/" rel="home">`)
header_frags[18] = []byte(`</a></li>
					`)
header_frags[19] = []byte(`
					<li class="menu_left menu_hamburger" title="`)
header_frags[20] = []byte(`"><a></a></li>
				</ul>
				</div>
				</div>
				<div style="clear: both;"></div>
			</nav>
			<div class="right_of_nav">`)
header_frags[21] = []byte(`</div>
			<div id="back"><div id="main" `)
header_frags[22] = []byte(`class="shrink_main"`)
header_frags[23] = []byte(`>
			<div class="alertbox">`)
notice_frags[0] = []byte(`<div class="alert">`)
notice_frags[1] = []byte(`</div>`)
header_frags[24] = []byte(`
			</div>
`)
topic_frags[0] = []byte(`

<form id="edit_topic_form" action='/topic/edit/submit/`)
topic_frags[1] = []byte(`?session=`)
topic_frags[2] = []byte(`' method="post"></form>
`)
topic_frags[3] = []byte(`<link rel="prev" href="/topic/`)
topic_frags[4] = []byte(`?page=`)
topic_frags[5] = []byte(`" />
<div id="prevFloat" class="prev_button"><a class="prev_link" aria-label="`)
topic_frags[6] = []byte(`" rel="prev" href="/topic/`)
topic_frags[7] = []byte(`?page=`)
topic_frags[8] = []byte(`">`)
topic_frags[9] = []byte(`</a></div>`)
topic_frags[10] = []byte(`<link rel="prerender next" href="/topic/`)
topic_frags[11] = []byte(`?page=`)
topic_frags[12] = []byte(`" />
<div id="nextFloat" class="next_button">
	<a class="next_link" aria-label="`)
topic_frags[13] = []byte(`" rel="next" href="/topic/`)
topic_frags[14] = []byte(`?page=`)
topic_frags[15] = []byte(`">`)
topic_frags[16] = []byte(`</a>
</div>`)
topic_frags[17] = []byte(`

<main>

<div  `)
topic_frags[18] = []byte(` class="rowblock rowhead topic_block" aria-label="`)
topic_frags[19] = []byte(`">
	<div class="rowitem topic_item`)
topic_frags[20] = []byte(` topic_sticky_head`)
topic_frags[21] = []byte(` topic_closed_head`)
topic_frags[22] = []byte(`">
		<h1 class='topic_name hide_on_edit' title='`)
topic_frags[23] = []byte(`'>`)
topic_frags[24] = []byte(`</h1>
		`)
topic_frags[25] = []byte(`<span class='username hide_on_micro topic_status_e topic_status_closed hide_on_edit' title='`)
topic_frags[26] = []byte(`' aria-label='`)
topic_frags[27] = []byte(`'>&#x1F512;&#xFE0E</span>`)
topic_frags[28] = []byte(`
		<input form='edit_topic_form' class='show_on_edit topic_name_input' name="topic_name" value='`)
topic_frags[29] = []byte(`' type="text" aria-label="`)
topic_frags[30] = []byte(`" />
		<button form='edit_topic_form' name="topic-button" class="formbutton show_on_edit submit_edit">`)
topic_frags[31] = []byte(`</button>
		`)
topic_frags[32] = []byte(`
	</div>
</div>
`)
topic_frags[33] = []byte(`
<article class="rowblock post_container poll" aria-level="`)
topic_frags[34] = []byte(`">
	<div class="rowitem passive editable_parent post_item poll_item `)
topic_frags[35] = []byte(`" style="background-image: url(`)
topic_frags[36] = []byte(`), url(/static/`)
topic_frags[37] = []byte(`/post-avatar-bg.jpg);background-position: 0px `)
topic_frags[38] = []byte(`-1`)
topic_frags[39] = []byte(`0px;background-repeat:no-repeat, repeat-y;">
		<div class="topic_content user_content" style="margin:0;padding:0;">
			`)
topic_frags[40] = []byte(`
			<div class="poll_option">
				<input form="poll_`)
topic_frags[41] = []byte(`_form" id="poll_option_`)
topic_frags[42] = []byte(`" name="poll_option_input" type="checkbox" value="`)
topic_frags[43] = []byte(`" />
				<label class="poll_option_label" for="poll_option_`)
topic_frags[44] = []byte(`">
					<div class="sel"></div>
				</label>
				<span id="poll_option_text_`)
topic_frags[45] = []byte(`" class="poll_option_text">`)
topic_frags[46] = []byte(`</span>
			</div>
			`)
topic_frags[47] = []byte(`
			<div class="poll_buttons">
				<button form="poll_`)
topic_frags[48] = []byte(`_form" class="poll_vote_button">`)
topic_frags[49] = []byte(`</button>
				<button class="poll_results_button" data-poll-id="`)
topic_frags[50] = []byte(`">`)
topic_frags[51] = []byte(`</button>
				<a href="#"><button class="poll_cancel_button">`)
topic_frags[52] = []byte(`</button></a>
			</div>
		</div>
		<div id="poll_results_`)
topic_frags[53] = []byte(`" class="poll_results auto_hide">
			<div class="topic_content user_content"></div>
		</div>
	</div>
</article>
`)
topic_frags[54] = []byte(`

<article `)
topic_frags[55] = []byte(` itemscope itemtype="http://schema.org/CreativeWork" class="rowblock post_container top_post" aria-label="`)
topic_frags[56] = []byte(`">
	<div class="rowitem passive editable_parent post_item `)
topic_frags[57] = []byte(`" style="background-image: url(`)
topic_frags[58] = []byte(`), url(/static/`)
topic_frags[59] = []byte(`/post-avatar-bg.jpg);background-position: 0px `)
topic_frags[60] = []byte(`-1`)
topic_frags[61] = []byte(`0px;background-repeat:no-repeat, repeat-y;">
		<p class="hide_on_edit topic_content user_content" itemprop="text" style="margin:0;padding:0;">`)
topic_frags[62] = []byte(`</p>
		<textarea name="topic_content" class="show_on_edit topic_content_input">`)
topic_frags[63] = []byte(`</textarea>

		<span class="controls`)
topic_frags[64] = []byte(` has_likes`)
topic_frags[65] = []byte(`" aria-label="`)
topic_frags[66] = []byte(`">

		<a href="`)
topic_frags[67] = []byte(`" class="username real_username" rel="author">`)
topic_frags[68] = []byte(`</a>&nbsp;&nbsp;
		`)
topic_frags[69] = []byte(`<a href="/topic/like/submit/`)
topic_frags[70] = []byte(`?session=`)
topic_frags[71] = []byte(`" class="mod_button"`)
topic_frags[72] = []byte(` title="`)
topic_frags[73] = []byte(`" aria-label="`)
topic_frags[74] = []byte(`"`)
topic_frags[75] = []byte(` title="`)
topic_frags[76] = []byte(`" aria-label="`)
topic_frags[77] = []byte(`"`)
topic_frags[78] = []byte(` style="color:#202020;">
		<button class="username like_label `)
topic_frags[79] = []byte(`remove_like`)
topic_frags[80] = []byte(`add_like`)
topic_frags[81] = []byte(`"></button></a>`)
topic_frags[82] = []byte(`<a href='/topic/edit/`)
topic_frags[83] = []byte(`' class="mod_button open_edit" style="font-weight:normal;" title="`)
topic_frags[84] = []byte(`" aria-label="`)
topic_frags[85] = []byte(`"><button class="username edit_label"></button></a>`)
topic_frags[86] = []byte(`<a href='/topic/delete/submit/`)
topic_frags[87] = []byte(`?session=`)
topic_frags[88] = []byte(`' class="mod_button" style="font-weight:normal;" title="`)
topic_frags[89] = []byte(`" aria-label="`)
topic_frags[90] = []byte(`"><button class="username trash_label"></button></a>`)
topic_frags[91] = []byte(`<a class="mod_button" href='/topic/unlock/submit/`)
topic_frags[92] = []byte(`?session=`)
topic_frags[93] = []byte(`' style="font-weight:normal;" title="`)
topic_frags[94] = []byte(`" aria-label="`)
topic_frags[95] = []byte(`"><button class="username unlock_label"></button></a>`)
topic_frags[96] = []byte(`<a href='/topic/lock/submit/`)
topic_frags[97] = []byte(`?session=`)
topic_frags[98] = []byte(`' class="mod_button" style="font-weight:normal;" title="`)
topic_frags[99] = []byte(`" aria-label="`)
topic_frags[100] = []byte(`"><button class="username lock_label"></button></a>`)
topic_frags[101] = []byte(`<a class="mod_button" href='/topic/unstick/submit/`)
topic_frags[102] = []byte(`?session=`)
topic_frags[103] = []byte(`' style="font-weight:normal;" title="`)
topic_frags[104] = []byte(`" aria-label="`)
topic_frags[105] = []byte(`"><button class="username unpin_label"></button></a>`)
topic_frags[106] = []byte(`<a href='/topic/stick/submit/`)
topic_frags[107] = []byte(`?session=`)
topic_frags[108] = []byte(`' class="mod_button" style="font-weight:normal;" title="`)
topic_frags[109] = []byte(`" aria-label="`)
topic_frags[110] = []byte(`"><button class="username pin_label"></button></a>`)
topic_frags[111] = []byte(`<a class="mod_button" href='/users/ips/?ip=`)
topic_frags[112] = []byte(`' style="font-weight:normal;" title="`)
topic_frags[113] = []byte(`" aria-label="The poster's IP is `)
topic_frags[114] = []byte(`"><button class="username ip_label"></button></a>`)
topic_frags[115] = []byte(`
		<a href="/report/submit/`)
topic_frags[116] = []byte(`?session=`)
topic_frags[117] = []byte(`&type=topic" class="mod_button report_item" style="font-weight:normal;" title="`)
topic_frags[118] = []byte(`" aria-label="`)
topic_frags[119] = []byte(`" rel="nofollow"><button class="username flag_label"></button></a>

		<a class="username hide_on_micro like_count" aria-label="`)
topic_frags[120] = []byte(`">`)
topic_frags[121] = []byte(`</a><a class="username hide_on_micro like_count_label" title="`)
topic_frags[122] = []byte(`"></a>

		`)
topic_frags[123] = []byte(`<a class="username hide_on_micro user_tag">`)
topic_frags[124] = []byte(`</a>`)
topic_frags[125] = []byte(`<a class="username hide_on_micro level" aria-label="`)
topic_frags[126] = []byte(`">`)
topic_frags[127] = []byte(`</a><a class="username hide_on_micro level_label" style="float:right;" title="`)
topic_frags[128] = []byte(`"></a>`)
topic_frags[129] = []byte(`

		</span>
	</div>
</article>

<div class="rowblock post_container" aria-label="`)
topic_frags[130] = []byte(`" style="overflow: hidden;">`)
topic_frags[131] = []byte(`
	<article itemscope itemtype="http://schema.org/CreativeWork" class="rowitem passive deletable_block editable_parent post_item action_item">
		<span class="action_icon" style="font-size: 18px;padding-right: 5px;">`)
topic_frags[132] = []byte(`</span>
		<span itemprop="text">`)
topic_frags[133] = []byte(`</span>
	</article>
`)
topic_frags[134] = []byte(`
	<article `)
topic_frags[135] = []byte(` itemscope itemtype="http://schema.org/CreativeWork" class="rowitem passive deletable_block editable_parent post_item `)
topic_frags[136] = []byte(`" style="background-image: url(`)
topic_frags[137] = []byte(`), url(/static/`)
topic_frags[138] = []byte(`/post-avatar-bg.jpg);background-position: 0px `)
topic_frags[139] = []byte(`-1`)
topic_frags[140] = []byte(`0px;background-repeat:no-repeat, repeat-y;">
		`)
topic_frags[141] = []byte(`
		<p class="editable_block user_content" itemprop="text" style="margin:0;padding:0;">`)
topic_frags[142] = []byte(`</p>

		<span class="controls`)
topic_frags[143] = []byte(` has_likes`)
topic_frags[144] = []byte(`">

		<a href="`)
topic_frags[145] = []byte(`" class="username real_username" rel="author">`)
topic_frags[146] = []byte(`</a>&nbsp;&nbsp;
		`)
topic_frags[147] = []byte(`<a href="/reply/like/submit/`)
topic_frags[148] = []byte(`?session=`)
topic_frags[149] = []byte(`" class="mod_button" title="`)
topic_frags[150] = []byte(`" aria-label="`)
topic_frags[151] = []byte(`" style="color:#202020;"><button class="username like_label remove_like"></button></a>`)
topic_frags[152] = []byte(`<a href="/reply/like/submit/`)
topic_frags[153] = []byte(`?session=`)
topic_frags[154] = []byte(`" class="mod_button" title="`)
topic_frags[155] = []byte(`" aria-label="`)
topic_frags[156] = []byte(`" style="color:#202020;"><button class="username like_label add_like"></button></a>`)
topic_frags[157] = []byte(`<a href="/reply/edit/submit/`)
topic_frags[158] = []byte(`?session=`)
topic_frags[159] = []byte(`" class="mod_button" title="`)
topic_frags[160] = []byte(`" aria-label="`)
topic_frags[161] = []byte(`"><button class="username edit_item edit_label"></button></a>`)
topic_frags[162] = []byte(`<a href="/reply/delete/submit/`)
topic_frags[163] = []byte(`?session=`)
topic_frags[164] = []byte(`" class="mod_button" title="`)
topic_frags[165] = []byte(`" aria-label="`)
topic_frags[166] = []byte(`"><button class="username delete_item trash_label"></button></a>`)
topic_frags[167] = []byte(`<a class="mod_button" href='/users/ips/?ip=`)
topic_frags[168] = []byte(`' style="font-weight:normal;" title="`)
topic_frags[169] = []byte(`" aria-label="The poster's IP is `)
topic_frags[170] = []byte(`"><button class="username ip_label"></button></a>`)
topic_frags[171] = []byte(`
		<a href="/report/submit/`)
topic_frags[172] = []byte(`?session=`)
topic_frags[173] = []byte(`&type=reply" class="mod_button report_item" title="`)
topic_frags[174] = []byte(`" aria-label="`)
topic_frags[175] = []byte(`" rel="nofollow"><button class="username report_item flag_label"></button></a>

		<a class="username hide_on_micro like_count">`)
topic_frags[176] = []byte(`</a><a class="username hide_on_micro like_count_label" title="`)
topic_frags[177] = []byte(`"></a>

		`)
topic_frags[178] = []byte(`<a class="username hide_on_micro user_tag">`)
topic_frags[179] = []byte(`</a>`)
topic_frags[180] = []byte(`<a class="username hide_on_micro level" aria-label="`)
topic_frags[181] = []byte(`">`)
topic_frags[182] = []byte(`</a><a class="username hide_on_micro level_label" style="float:right;" title="`)
topic_frags[183] = []byte(`"></a>`)
topic_frags[184] = []byte(`

		</span>
	</article>
`)
topic_frags[185] = []byte(`</div>

`)
topic_frags[186] = []byte(`
<div class="rowblock topic_reply_form quick_create_form" aria-label="`)
topic_frags[187] = []byte(`">
	<form id="quick_post_form" enctype="multipart/form-data" action="/reply/create/?session=`)
topic_frags[188] = []byte(`" method="post"></form>
	<input form="quick_post_form" name="tid" value='`)
topic_frags[189] = []byte(`' type="hidden" />
	<input form="quick_post_form" id="has_poll_input" name="has_poll" value="0" type="hidden" />
	<div class="formrow real_first_child">
		<div class="formitem">
			<textarea id="input_content" form="quick_post_form" name="reply-content" placeholder="`)
topic_frags[190] = []byte(`" required></textarea>
		</div>
	</div>
	<div class="formrow poll_content_row auto_hide">
		<div class="formitem">
			<div class="pollinput" data-pollinput="0">
				<input type="checkbox" disabled />
				<label class="pollinputlabel"></label>
				<input form="quick_post_form" name="pollinputitem[0]" class="pollinputinput" type="text" placeholder="`)
topic_frags[191] = []byte(`" />
			</div>
		</div>
	</div>
	<div class="formrow quick_button_row">
		<div class="formitem">
			<button form="quick_post_form" name="reply-button" class="formbutton">`)
topic_frags[192] = []byte(`</button>
			<button form="quick_post_form" class="formbutton" id="add_poll_button">`)
topic_frags[193] = []byte(`</button>
			`)
topic_frags[194] = []byte(`
			<input name="upload_files" form="quick_post_form" id="upload_files" multiple type="file" style="display: none;" />
			<label for="upload_files" class="formbutton add_file_button">`)
topic_frags[195] = []byte(`</label>
			<div id="upload_file_dock"></div>`)
topic_frags[196] = []byte(`
		</div>
	</div>
</div>
`)
topic_frags[197] = []byte(`

</main>

`)
footer_frags[0] = []byte(`<div class="footer">
	`)
footer_frags[1] = []byte(`
	<div id="poweredByHolder" class="footerBit">
		<div id="poweredBy">
			<a id="poweredByName" href="https://github.com/Azareal/Gosora">`)
footer_frags[2] = []byte(`</a><span id="poweredByDash"> - </span><span id="poweredByMaker">`)
footer_frags[3] = []byte(`</span>
		</div>
		<form action="/theme/" method="post">
			<div id="themeSelector" style="float: right;">
				<select id="themeSelectorSelect" name="themeSelector" aria-label="`)
footer_frags[4] = []byte(`">
				`)
footer_frags[5] = []byte(`<option val="`)
footer_frags[6] = []byte(`"`)
footer_frags[7] = []byte(` selected`)
footer_frags[8] = []byte(`>`)
footer_frags[9] = []byte(`</option>`)
footer_frags[10] = []byte(`
				</select>
			</div>
		</form>
	</div>
</div>
					</div>
				<aside class="sidebar">`)
footer_frags[11] = []byte(`</aside>
				<div style="clear: both;"></div>
			</div>
		</div>
	</body>
</html>
`)
topic_alt_frags[0] = []byte(`<link rel="prev" href="/topic/`)
topic_alt_frags[1] = []byte(`?page=`)
topic_alt_frags[2] = []byte(`" />
<div id="prevFloat" class="prev_button"><a class="prev_link" aria-label="`)
topic_alt_frags[3] = []byte(`" rel="prev" href="/topic/`)
topic_alt_frags[4] = []byte(`?page=`)
topic_alt_frags[5] = []byte(`">`)
topic_alt_frags[6] = []byte(`</a></div>`)
topic_alt_frags[7] = []byte(`<link rel="prerender next" href="/topic/`)
topic_alt_frags[8] = []byte(`?page=`)
topic_alt_frags[9] = []byte(`" />
<div id="nextFloat" class="next_button"><a class="next_link" aria-label="`)
topic_alt_frags[10] = []byte(`" rel="next" href="/topic/`)
topic_alt_frags[11] = []byte(`?page=`)
topic_alt_frags[12] = []byte(`">`)
topic_alt_frags[13] = []byte(`</a></div>`)
topic_alt_frags[14] = []byte(`

<main>

<div `)
topic_alt_frags[15] = []byte(` class="rowblock rowhead topic_block" aria-label="`)
topic_alt_frags[16] = []byte(`">
	<form action='/topic/edit/submit/`)
topic_alt_frags[17] = []byte(`?session=`)
topic_alt_frags[18] = []byte(`' method="post">
		<div class="rowitem topic_item`)
topic_alt_frags[19] = []byte(` topic_sticky_head`)
topic_alt_frags[20] = []byte(` topic_closed_head`)
topic_alt_frags[21] = []byte(`">
			<h1 class='topic_name hide_on_edit' title='`)
topic_alt_frags[22] = []byte(`'>`)
topic_alt_frags[23] = []byte(`</h1>
			`)
topic_alt_frags[24] = []byte(`<span class='username hide_on_micro topic_status_e topic_status_closed hide_on_edit' title='`)
topic_alt_frags[25] = []byte(`' aria-label='`)
topic_alt_frags[26] = []byte(`' style="font-weight:normal;float: right;position:relative;top:-5px;">&#x1F512;&#xFE0E</span>`)
topic_alt_frags[27] = []byte(`
			<input class='show_on_edit topic_name_input' name="topic_name" value='`)
topic_alt_frags[28] = []byte(`' type="text" aria-label="`)
topic_alt_frags[29] = []byte(`" />
			<button name="topic-button" class="formbutton show_on_edit submit_edit">`)
topic_alt_frags[30] = []byte(`</button>
			`)
topic_alt_frags[31] = []byte(`
		</div>
	</form>
</div>

<div class="rowblock post_container">
	`)
topic_alt_frags[32] = []byte(`
	<form id="poll_`)
topic_alt_frags[33] = []byte(`_form" action="/poll/vote/`)
topic_alt_frags[34] = []byte(`?session=`)
topic_alt_frags[35] = []byte(`" method="post"></form>
	<article class="rowitem passive deletable_block editable_parent post_item poll_item top_post hide_on_edit">
		<div class="userinfo" aria-label="`)
topic_alt_frags[36] = []byte(`">
			<div class="avatar_item" style="background-image: url(`)
topic_alt_frags[37] = []byte(`), url(/static/white-dot.jpg);background-position: 0px -10px;">&nbsp;</div>
			<a href="`)
topic_alt_frags[38] = []byte(`" class="the_name" rel="author">`)
topic_alt_frags[39] = []byte(`</a>
			`)
topic_alt_frags[40] = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag">`)
topic_alt_frags[41] = []byte(`</div><div class="tag_post"></div></div>`)
topic_alt_frags[42] = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag post_level">`)
topic_alt_frags[43] = []byte(`</div><div class="tag_post"></div></div>`)
topic_alt_frags[44] = []byte(`
		</div>
		<div id="poll_voter_`)
topic_alt_frags[45] = []byte(`" class="content_container poll_voter">
			<div class="topic_content user_content">
				`)
topic_alt_frags[46] = []byte(`
				<div class="poll_option">
					<input form="poll_`)
topic_alt_frags[47] = []byte(`_form" id="poll_option_`)
topic_alt_frags[48] = []byte(`" name="poll_option_input" type="checkbox" value="`)
topic_alt_frags[49] = []byte(`" />
					<label class="poll_option_label" for="poll_option_`)
topic_alt_frags[50] = []byte(`">
						<div class="sel"></div>
					</label>
					<span id="poll_option_text_`)
topic_alt_frags[51] = []byte(`" class="poll_option_text">`)
topic_alt_frags[52] = []byte(`</span>
				</div>
				`)
topic_alt_frags[53] = []byte(`
				<div class="poll_buttons">
					<button form="poll_`)
topic_alt_frags[54] = []byte(`_form" class="poll_vote_button">`)
topic_alt_frags[55] = []byte(`</button>
					<button class="poll_results_button" data-poll-id="`)
topic_alt_frags[56] = []byte(`">`)
topic_alt_frags[57] = []byte(`</button>
					<a href="#"><button class="poll_cancel_button">`)
topic_alt_frags[58] = []byte(`</button></a>
				</div>
			</div>
		</div>
		<div id="poll_results_`)
topic_alt_frags[59] = []byte(`" class="content_container poll_results auto_hide">
			<div class="topic_content user_content"></div>
		</div>
	</article>
	`)
topic_alt_frags[60] = []byte(`
	<article `)
topic_alt_frags[61] = []byte(` itemscope itemtype="http://schema.org/CreativeWork" class="rowitem passive deletable_block editable_parent post_item top_post" aria-label="`)
topic_alt_frags[62] = []byte(`">
		<div class="userinfo" aria-label="`)
topic_alt_frags[63] = []byte(`">
			<div class="avatar_item" style="background-image: url(`)
topic_alt_frags[64] = []byte(`), url(/static/white-dot.jpg);background-position: 0px -10px;">&nbsp;</div>
			<a href="`)
topic_alt_frags[65] = []byte(`" class="the_name" rel="author">`)
topic_alt_frags[66] = []byte(`</a>
			`)
topic_alt_frags[67] = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag">`)
topic_alt_frags[68] = []byte(`</div><div class="tag_post"></div></div>`)
topic_alt_frags[69] = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag post_level">`)
topic_alt_frags[70] = []byte(`</div><div class="tag_post"></div></div>`)
topic_alt_frags[71] = []byte(`
		</div>
		<div class="content_container">
			<div class="hide_on_edit topic_content user_content" itemprop="text">`)
topic_alt_frags[72] = []byte(`</div>
			<textarea name="topic_content" class="show_on_edit topic_content_input">`)
topic_alt_frags[73] = []byte(`</textarea>
			<div class="controls button_container`)
topic_alt_frags[74] = []byte(` has_likes`)
topic_alt_frags[75] = []byte(`">
				`)
topic_alt_frags[76] = []byte(`<a href="/topic/like/submit/`)
topic_alt_frags[77] = []byte(`?session=`)
topic_alt_frags[78] = []byte(`" class="action_button like_item `)
topic_alt_frags[79] = []byte(`remove_like`)
topic_alt_frags[80] = []byte(`add_like`)
topic_alt_frags[81] = []byte(`" aria-label="`)
topic_alt_frags[82] = []byte(`" data-action="like"></a>`)
topic_alt_frags[83] = []byte(`<a href="/topic/edit/`)
topic_alt_frags[84] = []byte(`" class="action_button open_edit" aria-label="`)
topic_alt_frags[85] = []byte(`" data-action="edit"></a>`)
topic_alt_frags[86] = []byte(`<a href="/topic/delete/submit/`)
topic_alt_frags[87] = []byte(`?session=`)
topic_alt_frags[88] = []byte(`" class="action_button delete_item" aria-label="`)
topic_alt_frags[89] = []byte(`" data-action="delete"></a>`)
topic_alt_frags[90] = []byte(`<a href='/topic/unlock/submit/`)
topic_alt_frags[91] = []byte(`?session=`)
topic_alt_frags[92] = []byte(`' class="action_button unlock_item" data-action="unlock" aria-label="`)
topic_alt_frags[93] = []byte(`"></a>`)
topic_alt_frags[94] = []byte(`<a href='/topic/lock/submit/`)
topic_alt_frags[95] = []byte(`?session=`)
topic_alt_frags[96] = []byte(`' class="action_button lock_item" data-action="lock" aria-label="`)
topic_alt_frags[97] = []byte(`"></a>`)
topic_alt_frags[98] = []byte(`<a href='/topic/unstick/submit/`)
topic_alt_frags[99] = []byte(`?session=`)
topic_alt_frags[100] = []byte(`' class="action_button unpin_item" data-action="unpin" aria-label="`)
topic_alt_frags[101] = []byte(`"></a>`)
topic_alt_frags[102] = []byte(`<a href='/topic/stick/submit/`)
topic_alt_frags[103] = []byte(`?session=`)
topic_alt_frags[104] = []byte(`' class="action_button pin_item" data-action="pin" aria-label="`)
topic_alt_frags[105] = []byte(`"></a>`)
topic_alt_frags[106] = []byte(`<a href="/users/ips/?ip=`)
topic_alt_frags[107] = []byte(`" title="`)
topic_alt_frags[108] = []byte(`" class="action_button ip_item_button hide_on_big" aria-label="`)
topic_alt_frags[109] = []byte(`" data-action="ip"></a>`)
topic_alt_frags[110] = []byte(`
					<a href="/report/submit/`)
topic_alt_frags[111] = []byte(`?session=`)
topic_alt_frags[112] = []byte(`&type=topic" class="action_button report_item" aria-label="`)
topic_alt_frags[113] = []byte(`" data-action="report"></a>
					<a href="#" class="action_button button_menu"></a>
				`)
topic_alt_frags[114] = []byte(`
				<div class="action_button_right">
					<a class="action_button like_count hide_on_micro" aria-label="`)
topic_alt_frags[115] = []byte(`">`)
topic_alt_frags[116] = []byte(`</a>
					<a class="action_button created_at hide_on_mobile">`)
topic_alt_frags[117] = []byte(`</a>
					`)
topic_alt_frags[118] = []byte(`<a href="/users/ips/?ip=`)
topic_alt_frags[119] = []byte(`" title="`)
topic_alt_frags[120] = []byte(`" class="action_button ip_item hide_on_mobile" aria-hidden="true">`)
topic_alt_frags[121] = []byte(`</a>`)
topic_alt_frags[122] = []byte(`
				</div>
			</div>
		</div><div style="clear:both;"></div>
	</article>

	`)
topic_alt_frags[123] = []byte(`
	<article `)
topic_alt_frags[124] = []byte(` itemscope itemtype="http://schema.org/CreativeWork" class="rowitem passive deletable_block editable_parent post_item `)
topic_alt_frags[125] = []byte(`action_item`)
topic_alt_frags[126] = []byte(`">
		<div class="userinfo" aria-label="`)
topic_alt_frags[127] = []byte(`">
			<div class="avatar_item" style="background-image: url(`)
topic_alt_frags[128] = []byte(`), url(/static/white-dot.jpg);background-position: 0px -10px;">&nbsp;</div>
			<a href="`)
topic_alt_frags[129] = []byte(`" class="the_name" rel="author">`)
topic_alt_frags[130] = []byte(`</a>
			`)
topic_alt_frags[131] = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag">`)
topic_alt_frags[132] = []byte(`</div><div class="tag_post"></div></div>`)
topic_alt_frags[133] = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag post_level">`)
topic_alt_frags[134] = []byte(`</div><div class="tag_post"></div></div>`)
topic_alt_frags[135] = []byte(`
		</div>
		<div class="content_container" `)
topic_alt_frags[136] = []byte(`style="margin-left: 0px;"`)
topic_alt_frags[137] = []byte(`>
			`)
topic_alt_frags[138] = []byte(`
				<span class="action_icon" style="font-size: 18px;padding-right: 5px;" aria-hidden="true">`)
topic_alt_frags[139] = []byte(`</span>
				<span itemprop="text">`)
topic_alt_frags[140] = []byte(`</span>
			`)
topic_alt_frags[141] = []byte(`
			<div class="editable_block user_content" itemprop="text">`)
topic_alt_frags[142] = []byte(`</div>
			<div class="controls button_container`)
topic_alt_frags[143] = []byte(` has_likes`)
topic_alt_frags[144] = []byte(`">
				`)
topic_alt_frags[145] = []byte(`<a href="/reply/like/submit/`)
topic_alt_frags[146] = []byte(`?session=`)
topic_alt_frags[147] = []byte(`" class="action_button like_item `)
topic_alt_frags[148] = []byte(`remove_like`)
topic_alt_frags[149] = []byte(`add_like`)
topic_alt_frags[150] = []byte(`" aria-label="`)
topic_alt_frags[151] = []byte(`" data-action="like"></a>`)
topic_alt_frags[152] = []byte(`<a href="/reply/edit/submit/`)
topic_alt_frags[153] = []byte(`?session=`)
topic_alt_frags[154] = []byte(`" class="action_button edit_item" aria-label="`)
topic_alt_frags[155] = []byte(`" data-action="edit"></a>`)
topic_alt_frags[156] = []byte(`<a href="/reply/delete/submit/`)
topic_alt_frags[157] = []byte(`?session=`)
topic_alt_frags[158] = []byte(`" class="action_button delete_item" aria-label="`)
topic_alt_frags[159] = []byte(`" data-action="delete"></a>`)
topic_alt_frags[160] = []byte(`<a href="/users/ips/?ip=`)
topic_alt_frags[161] = []byte(`" title="`)
topic_alt_frags[162] = []byte(`" class="action_button ip_item_button hide_on_big" aria-label="`)
topic_alt_frags[163] = []byte(`" data-action="ip"></a>`)
topic_alt_frags[164] = []byte(`
					<a href="/report/submit/`)
topic_alt_frags[165] = []byte(`?session=`)
topic_alt_frags[166] = []byte(`&type=reply" class="action_button report_item" aria-label="`)
topic_alt_frags[167] = []byte(`" data-action="report"></a>
					<a href="#" class="action_button button_menu"></a>
				`)
topic_alt_frags[168] = []byte(`
				<div class="action_button_right">
					<a class="action_button like_count hide_on_micro" aria-label="`)
topic_alt_frags[169] = []byte(`">`)
topic_alt_frags[170] = []byte(`</a>
					<a class="action_button created_at hide_on_mobile">`)
topic_alt_frags[171] = []byte(`</a>
					`)
topic_alt_frags[172] = []byte(`<a href="/users/ips/?ip=`)
topic_alt_frags[173] = []byte(`" title="IP Address" class="action_button ip_item hide_on_mobile" aria-hidden="true">`)
topic_alt_frags[174] = []byte(`</a>`)
topic_alt_frags[175] = []byte(`
				</div>
			</div>
			`)
topic_alt_frags[176] = []byte(`
		</div>
		<div style="clear:both;"></div>
	</article>
`)
topic_alt_frags[177] = []byte(`</div>

`)
topic_alt_frags[178] = []byte(`
<div class="rowblock topic_reply_container">
	<div class="userinfo" aria-label="`)
topic_alt_frags[179] = []byte(`">
		<div class="avatar_item" style="background-image: url(`)
topic_alt_frags[180] = []byte(`), url(/static/white-dot.jpg);background-position: 0px -10px;">&nbsp;</div>
		<a href="`)
topic_alt_frags[181] = []byte(`" class="the_name" rel="author">`)
topic_alt_frags[182] = []byte(`</a>
		`)
topic_alt_frags[183] = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag">`)
topic_alt_frags[184] = []byte(`</div><div class="tag_post"></div></div>`)
topic_alt_frags[185] = []byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag post_level">`)
topic_alt_frags[186] = []byte(`</div><div class="tag_post"></div></div>`)
topic_alt_frags[187] = []byte(`
	</div>
	<div class="rowblock topic_reply_form quick_create_form"  aria-label="`)
topic_alt_frags[188] = []byte(`">
		<form id="quick_post_form" enctype="multipart/form-data" action="/reply/create/?session=`)
topic_alt_frags[189] = []byte(`" method="post"></form>
		<input form="quick_post_form" name="tid" value='`)
topic_alt_frags[190] = []byte(`' type="hidden" />
		<input form="quick_post_form" id="has_poll_input" name="has_poll" value="0" type="hidden" />
		<div class="formrow real_first_child">
			<div class="formitem">
				<textarea id="input_content" form="quick_post_form" name="reply-content" placeholder="`)
topic_alt_frags[191] = []byte(`" required></textarea>
			</div>
		</div>
		<div class="formrow poll_content_row auto_hide">
			<div class="formitem">
				<div class="pollinput" data-pollinput="0">
					<input type="checkbox" disabled />
					<label class="pollinputlabel"></label>
					<input form="quick_post_form" name="pollinputitem[0]" class="pollinputinput" type="text" placeholder="`)
topic_alt_frags[192] = []byte(`" />
				</div>
			</div>
		</div>
		<div class="formrow quick_button_row">
			<div class="formitem">
				<button form="quick_post_form" name="reply-button" class="formbutton">`)
topic_alt_frags[193] = []byte(`</button>
				<button form="quick_post_form" class="formbutton" id="add_poll_button">`)
topic_alt_frags[194] = []byte(`</button>
				`)
topic_alt_frags[195] = []byte(`
				<input name="upload_files" form="quick_post_form" id="upload_files" multiple type="file" style="display: none;" />
				<label for="upload_files" class="formbutton add_file_button">`)
topic_alt_frags[196] = []byte(`</label>
				<div id="upload_file_dock"></div>`)
topic_alt_frags[197] = []byte(`
			</div>
		</div>
	</div>
</div>
`)
topic_alt_frags[198] = []byte(`

</main>

`)
profile_frags[0] = []byte(`

<div id="profile_container" class="colstack">

<div id="profile_left_lane" class="colstack_left">
	<div id="profile_left_pane" class="rowmenu">
		<div class="topBlock">
			<div class="rowitem avatarRow">
				<img src="`)
profile_frags[1] = []byte(`" class="avatar" alt="`)
profile_frags[2] = []byte(`'s Avatar" title="`)
profile_frags[3] = []byte(`'s Avatar" />
			</div>
			<div class="rowitem nameRow">
				<span class="profileName" title="`)
profile_frags[4] = []byte(`">`)
profile_frags[5] = []byte(`</span>`)
profile_frags[6] = []byte(`<span class="username" title="`)
profile_frags[7] = []byte(`">`)
profile_frags[8] = []byte(`</span>`)
profile_frags[9] = []byte(`
			</div>
		</div>
		<div class="passiveBlock">
			`)
profile_frags[10] = []byte(`<div class="rowitem passive">
				<a class="profile_menu_item">`)
profile_frags[11] = []byte(`</a>
			</div>`)
profile_frags[12] = []byte(`
			<!--<div class="rowitem passive">
				<a class="profile_menu_item">`)
profile_frags[13] = []byte(`</a>
			</div>-->
			`)
profile_frags[14] = []byte(`<div class="rowitem passive">
				`)
profile_frags[15] = []byte(`<a href="/users/unban/`)
profile_frags[16] = []byte(`?session=`)
profile_frags[17] = []byte(`" class="profile_menu_item">`)
profile_frags[18] = []byte(`</a>
			`)
profile_frags[19] = []byte(`<a href="#ban_user" class="profile_menu_item">`)
profile_frags[20] = []byte(`</a>`)
profile_frags[21] = []byte(`
			</div>`)
profile_frags[22] = []byte(`
			<div class="rowitem passive">
				<a href="/report/submit/`)
profile_frags[23] = []byte(`?session=`)
profile_frags[24] = []byte(`&type=user" class="profile_menu_item report_item" aria-label="`)
profile_frags[25] = []byte(`" title="`)
profile_frags[26] = []byte(`"></a>
			</div>
			`)
profile_frags[27] = []byte(`
		</div>
	</div>
</div>

<div id="profile_right_lane" class="colstack_right">
	`)
profile_frags[28] = []byte(`
	<!-- TODO: Inline the display: none; CSS -->
	<div id="ban_user_head" class="colstack_item colstack_head hash_hide ban_user_hash" style="display: none;">
			<div class="rowitem"><h1><a>`)
profile_frags[29] = []byte(`</a></h1></div>
	</div>
	<form id="ban_user_form" class="hash_hide ban_user_hash" action="/users/ban/submit/`)
profile_frags[30] = []byte(`?session=`)
profile_frags[31] = []byte(`" method="post" style="display: none;">
		`)
profile_frags[32] = []byte(`
		<div class="colline">`)
profile_frags[33] = []byte(`</div>
		<div class="colstack_item">
			<div class="formrow real_first_child">
				<div class="formitem formlabel"><a>`)
profile_frags[34] = []byte(`</a></div>
				<div class="formitem">
					<input name="ban-duration-days" type="number" value="0" min="0" />
				</div>
			</div>
			<div class="formrow">
				<div class="formitem formlabel"><a>`)
profile_frags[35] = []byte(`</a></div>
				<div class="formitem">
					<input name="ban-duration-weeks" type="number" value="0" min="0" />
				</div>
			</div>
			<div class="formrow">
				<div class="formitem formlabel"><a>`)
profile_frags[36] = []byte(`</a></div>
				<div class="formitem">
					<input name="ban-duration-months" type="number" value="0" min="0" />
				</div>
			</div>
			<!--<div class="formrow">
				<div class="formitem formlabel"><a>`)
profile_frags[37] = []byte(`</a></div>
				<div class="formitem"><textarea name="ban-reason" placeholder="A really horrible person" required></textarea></div>
			</div>-->
			<div class="formrow">
				<div class="formitem"><button name="ban-button" class="formbutton form_middle_button">`)
profile_frags[38] = []byte(`</button></div>
			</div>
		</div>
	</form>
	`)
profile_frags[39] = []byte(`

	<div id="profile_comments_head" class="colstack_item colstack_head hash_hide">
		<div class="rowitem"><h1><a>`)
profile_frags[40] = []byte(`</a></h1></div>
	</div>
	<div id="profile_comments" class="colstack_item hash_hide">`)
profile_comments_row_frags[0] = []byte(`
		<div class="rowitem passive deletable_block editable_parent simple `)
profile_comments_row_frags[1] = []byte(`" style="background-image: url(`)
profile_comments_row_frags[2] = []byte(`), url(/static/post-avatar-bg.jpg);background-position: 0px `)
profile_comments_row_frags[3] = []byte(`-1`)
profile_comments_row_frags[4] = []byte(`0px;">
			<span class="editable_block user_content simple">`)
profile_comments_row_frags[5] = []byte(`</span>
			<span class="controls">
				<a href="`)
profile_comments_row_frags[6] = []byte(`" class="real_username username">`)
profile_comments_row_frags[7] = []byte(`</a>&nbsp;&nbsp;

				`)
profile_comments_row_frags[8] = []byte(`<a href="/profile/reply/edit/submit/`)
profile_comments_row_frags[9] = []byte(`?session=`)
profile_comments_row_frags[10] = []byte(`" class="mod_button" title="`)
profile_comments_row_frags[11] = []byte(`" aria-label="`)
profile_comments_row_frags[12] = []byte(`"><button class="username edit_item edit_label"></button></a>

				<a href="/profile/reply/delete/submit/`)
profile_comments_row_frags[13] = []byte(`?session=`)
profile_comments_row_frags[14] = []byte(`" class="mod_button" title="`)
profile_comments_row_frags[15] = []byte(`" aria-label="`)
profile_comments_row_frags[16] = []byte(`"><button class="username delete_item trash_label"></button></a>`)
profile_comments_row_frags[17] = []byte(`

				<a class="mod_button" href="/report/submit/`)
profile_comments_row_frags[18] = []byte(`?session=`)
profile_comments_row_frags[19] = []byte(`&type=user-reply"><button class="username report_item flag_label" title="`)
profile_comments_row_frags[20] = []byte(`" aria-label="`)
profile_comments_row_frags[21] = []byte(`"></button></a>

				`)
profile_comments_row_frags[22] = []byte(`<a class="username hide_on_mobile user_tag" style="float: right;">`)
profile_comments_row_frags[23] = []byte(`</a>`)
profile_comments_row_frags[24] = []byte(`
			</span>
		</div>
	`)
profile_comments_row_frags[25] = []byte(`
		<div class="rowitem passive deletable_block editable_parent comment `)
profile_comments_row_frags[26] = []byte(`">
			<div class="topRow">
				<div class="userbit">
					<img src="`)
profile_comments_row_frags[27] = []byte(`" alt="`)
profile_comments_row_frags[28] = []byte(`'s Avatar" title="`)
profile_comments_row_frags[29] = []byte(`'s Avatar" />
					<span class="nameAndTitle">
						<a href="`)
profile_comments_row_frags[30] = []byte(`" class="real_username username">`)
profile_comments_row_frags[31] = []byte(`</a>
						`)
profile_comments_row_frags[32] = []byte(`<a class="username hide_on_mobile user_tag" style="float: right;">`)
profile_comments_row_frags[33] = []byte(`</a>`)
profile_comments_row_frags[34] = []byte(`
					</span>
				</div>
				<span class="controls">
					`)
profile_comments_row_frags[35] = []byte(`
						<a href="/profile/reply/edit/submit/`)
profile_comments_row_frags[36] = []byte(`?session=`)
profile_comments_row_frags[37] = []byte(`" class="mod_button" title="`)
profile_comments_row_frags[38] = []byte(`" aria-label="`)
profile_comments_row_frags[39] = []byte(`"><button class="username edit_item edit_label"></button></a>
						<a href="/profile/reply/delete/submit/`)
profile_comments_row_frags[40] = []byte(`?session=`)
profile_comments_row_frags[41] = []byte(`" class="mod_button" title="`)
profile_comments_row_frags[42] = []byte(`" aria-label="`)
profile_comments_row_frags[43] = []byte(`"><button class="username delete_item trash_label"></button></a>
					`)
profile_comments_row_frags[44] = []byte(`
					<a class="mod_button" href="/report/submit/`)
profile_comments_row_frags[45] = []byte(`?session=`)
profile_comments_row_frags[46] = []byte(`&type=user-reply"><button class="username report_item flag_label" title="`)
profile_comments_row_frags[47] = []byte(`" aria-label="`)
profile_comments_row_frags[48] = []byte(`"></button></a>
				</span>
			</div>
			<div class="content_column">
				<span class="editable_block user_content">`)
profile_comments_row_frags[49] = []byte(`</span>
			</div>
		</div>
		<div class="after_comment"></div>
	`)
profile_frags[41] = []byte(`</div>

`)
profile_frags[42] = []byte(`
	<form id="profile_comments_form" class="hash_hide" action="/profile/reply/create/?session=`)
profile_frags[43] = []byte(`" method="post">
		<input name="uid" value='`)
profile_frags[44] = []byte(`' type="hidden" />
		<div class="colstack_item topic_reply_form" style="border-top: none;">
			<div class="formrow">
				<div class="formitem"><textarea class="input_content" name="reply-content" placeholder="`)
profile_frags[45] = []byte(`"></textarea></div>
			</div>
			<div class="formrow quick_button_row">
				<div class="formitem"><button name="reply-button" class="formbutton">`)
profile_frags[46] = []byte(`</button></div>
			</div>
		</div>
	</form>
`)
profile_frags[47] = []byte(`
</div>

</div>

`)
profile_frags[48] = []byte(`
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
forums_frags[0] = []byte(`
<main id="forumsItemList" itemscope itemtype="http://schema.org/ItemList">

<div class="rowblock opthead">
	<div class="rowitem"><h1 itemprop="name">`)
forums_frags[1] = []byte(`</h1></div>
</div>
<div class="rowblock forum_list">
	`)
forums_frags[2] = []byte(`<div class="rowitem `)
forums_frags[3] = []byte(`datarow `)
forums_frags[4] = []byte(`"itemprop="itemListElement" itemscope
      itemtype="http://schema.org/ListItem">
		<span class="forum_left shift_left">
			<a href="`)
forums_frags[5] = []byte(`" itemprop="item">`)
forums_frags[6] = []byte(`</a>
		`)
forums_frags[7] = []byte(`
			<br /><span class="rowsmall" itemprop="description">`)
forums_frags[8] = []byte(`</span>
		`)
forums_frags[9] = []byte(`
			<br /><span class="rowsmall" style="font-style: italic;">`)
forums_frags[10] = []byte(`</span>
		`)
forums_frags[11] = []byte(`
		</span>

		<span class="forum_right shift_right">
			`)
forums_frags[12] = []byte(`<img class="extra_little_row_avatar" src="`)
forums_frags[13] = []byte(`" height=64 width=64 alt="`)
forums_frags[14] = []byte(`'s Avatar" title="`)
forums_frags[15] = []byte(`'s Avatar" />`)
forums_frags[16] = []byte(`
			<span>
				<a href="`)
forums_frags[17] = []byte(`">`)
forums_frags[18] = []byte(`</a>
				`)
forums_frags[19] = []byte(`<br /><span class="rowsmall">`)
forums_frags[20] = []byte(`</span>`)
forums_frags[21] = []byte(`
			</span>
		</span>
		<div style="clear: both;"></div>
	</div>
	`)
forums_frags[22] = []byte(`<div class="rowitem passive rowmsg">`)
forums_frags[23] = []byte(`</div>`)
forums_frags[24] = []byte(`
</div>

</main>
`)
topics_frags[0] = []byte(`
<main id="topicsItemList" itemscope itemtype="http://schema.org/ItemList">

<div class="rowblock rowhead topic_list_title_block`)
topics_frags[1] = []byte(` has_opt`)
topics_frags[2] = []byte(`">
	<div class="rowitem topic_list_title"><h1 itemprop="name">`)
topics_frags[3] = []byte(`</h1></div>
	`)
topics_frags[4] = []byte(`
		<div class="optbox">
		`)
topics_frags[5] = []byte(`
			<div class="pre_opt auto_hide"></div>
			<div class="opt create_topic_opt" title="`)
topics_frags[6] = []byte(`" aria-label="`)
topics_frags[7] = []byte(`"><a class="create_topic_link" href="/topics/create/"></a></div>
			`)
topics_frags[8] = []byte(`
			<div class="opt mod_opt" title="`)
topics_frags[9] = []byte(`">
				<a class="moderate_link" href="#" aria-label="`)
topics_frags[10] = []byte(`"></a>
			</div>
			`)
topics_frags[11] = []byte(`<div class="opt locked_opt" title="`)
topics_frags[12] = []byte(`" aria-label="`)
topics_frags[13] = []byte(`"><a></a></div>`)
topics_frags[14] = []byte(`
		</div>
		<div style="clear: both;"></div>
	`)
topics_frags[15] = []byte(`
</div>

`)
topics_frags[16] = []byte(`
<div class="mod_floater auto_hide">
	<form method="post">
	<div class="mod_floater_head">
		<span>`)
topics_frags[17] = []byte(`</span>
	</div>
	<div class="mod_floater_body">
		<select class="mod_floater_options">
			<option val="delete">`)
topics_frags[18] = []byte(`</option>
			<option val="lock">`)
topics_frags[19] = []byte(`</option>
			<option val="move">`)
topics_frags[20] = []byte(`</option>
		</select>
		<button class="mod_floater_submit">`)
topics_frags[21] = []byte(`</button>
	</div>
	</form>
</div>

`)
topics_frags[22] = []byte(`
<div id="mod_topic_mover" class="modal_pane auto_hide">
	<form action="/topic/move/submit/?session=`)
topics_frags[23] = []byte(`" method="post">
		<input id="mover_fid" name="fid" value="0" type="hidden" />
		<div class="pane_header">
			<h3>`)
topics_frags[24] = []byte(`</h3>
		</div>
		<div class="pane_body">
			<div class="pane_table">
				`)
topics_frags[25] = []byte(`<div id="mover_fid_`)
topics_frags[26] = []byte(`" data-fid="`)
topics_frags[27] = []byte(`" class="pane_row">`)
topics_frags[28] = []byte(`</div>`)
topics_frags[29] = []byte(`
			</div>
		</div>
		<div class="pane_buttons">
			<button id="mover_submit">`)
topics_frags[30] = []byte(`</button>
		</div>
	</form>
</div>
<div class="rowblock topic_create_form quick_create_form" style="display: none;" aria-label="`)
topics_frags[31] = []byte(`">
	<form name="topic_create_form_form" id="quick_post_form" enctype="multipart/form-data" action="/topic/create/submit/?session=`)
topics_frags[32] = []byte(`" method="post"></form>
	<input form="quick_post_form" id="has_poll_input" name="has_poll" value="0" type="hidden" />
	<img class="little_row_avatar" src="`)
topics_frags[33] = []byte(`" height="64" alt="`)
topics_frags[34] = []byte(`" title="`)
topics_frags[35] = []byte(`" />
	<div class="main_form">
		<div class="topic_meta">
			<div class="formrow topic_board_row real_first_child">
				<div class="formitem"><select form="quick_post_form" id="topic_board_input" name="topic-board">
					`)
topics_frags[36] = []byte(`<option `)
topics_frags[37] = []byte(`selected`)
topics_frags[38] = []byte(` value="`)
topics_frags[39] = []byte(`">`)
topics_frags[40] = []byte(`</option>`)
topics_frags[41] = []byte(`
				</select></div>
			</div>
			<div class="formrow topic_name_row">
				<div class="formitem">
					<input form="quick_post_form" name="topic-name" placeholder="`)
topics_frags[42] = []byte(`" required>
				</div>
			</div>
		</div>
		<div class="formrow topic_content_row">
			<div class="formitem">
				<textarea form="quick_post_form" id="input_content" name="topic-content" placeholder="`)
topics_frags[43] = []byte(`" required></textarea>
			</div>
		</div>
		<div class="formrow poll_content_row auto_hide">
			<div class="formitem">
				<div class="pollinput" data-pollinput="0">
					<input type="checkbox" disabled />
					<label class="pollinputlabel"></label>
					<input form="quick_post_form" name="pollinputitem[0]" class="pollinputinput" type="text" placeholder="`)
topics_frags[44] = []byte(`" />
				</div>
			</div>
		</div>
		<div class="formrow quick_button_row">
			<div class="formitem">
				<button form="quick_post_form" class="formbutton">`)
topics_frags[45] = []byte(`</button>
				<button form="quick_post_form" class="formbutton" id="add_poll_button">`)
topics_frags[46] = []byte(`</button>
				`)
topics_frags[47] = []byte(`
				<input name="upload_files" form="quick_post_form" id="upload_files" multiple type="file" style="display: none;" />
				<label for="upload_files" class="formbutton add_file_button">`)
topics_frags[48] = []byte(`</label>
				<div id="upload_file_dock"></div>`)
topics_frags[49] = []byte(`
				<button class="formbutton close_form">`)
topics_frags[50] = []byte(`</button>
			</div>
		</div>
	</div>
</div>
	`)
topics_frags[51] = []byte(`
<div id="topic_list" class="rowblock topic_list" aria-label="`)
topics_frags[52] = []byte(`">
	`)
topics_frags[53] = []byte(`<div class="topic_row" data-tid="`)
topics_frags[54] = []byte(`">
	<div class="rowitem topic_left passive datarow `)
topics_frags[55] = []byte(`topic_sticky`)
topics_frags[56] = []byte(`topic_closed`)
topics_frags[57] = []byte(`">
		<span class="selector"></span>
		<a href="`)
topics_frags[58] = []byte(`"><img src="`)
topics_frags[59] = []byte(`" height="64" alt="`)
topics_frags[60] = []byte(`'s Avatar" title="`)
topics_frags[61] = []byte(`'s Avatar" /></a>
		<span class="topic_inner_left">
			<a class="rowtopic" href="`)
topics_frags[62] = []byte(`" itemprop="itemListElement" title="`)
topics_frags[63] = []byte(`"><span>`)
topics_frags[64] = []byte(`</span></a> `)
topics_frags[65] = []byte(`<a class="rowsmall parent_forum" href="`)
topics_frags[66] = []byte(`" title="`)
topics_frags[67] = []byte(`">`)
topics_frags[68] = []byte(`</a>`)
topics_frags[69] = []byte(`
			<br /><a class="rowsmall starter" href="`)
topics_frags[70] = []byte(`" title="`)
topics_frags[71] = []byte(`">`)
topics_frags[72] = []byte(`</a>
			`)
topics_frags[73] = []byte(`<span class="rowsmall topic_status_e topic_status_closed" title="`)
topics_frags[74] = []byte(`"> | &#x1F512;&#xFE0E</span>`)
topics_frags[75] = []byte(`<span class="rowsmall topic_status_e topic_status_sticky" title="`)
topics_frags[76] = []byte(`"> | &#x1F4CD;&#xFE0E</span>`)
topics_frags[77] = []byte(`
		</span>
		<span class="topic_inner_right rowsmall" style="float: right;">
			<span class="replyCount">`)
topics_frags[78] = []byte(`</span><br />
			<span class="likeCount">`)
topics_frags[79] = []byte(`</span>
		</span>
	</div>
	<div class="rowitem topic_right passive datarow `)
topics_frags[80] = []byte(`topic_sticky`)
topics_frags[81] = []byte(`topic_closed`)
topics_frags[82] = []byte(`">
		<a href="`)
topics_frags[83] = []byte(`"><img src="`)
topics_frags[84] = []byte(`" height="64" alt="`)
topics_frags[85] = []byte(`'s Avatar" title="`)
topics_frags[86] = []byte(`'s Avatar" /></a>
		<span>
			<a href="`)
topics_frags[87] = []byte(`" class="lastName" style="font-size: 14px;" title="`)
topics_frags[88] = []byte(`">`)
topics_frags[89] = []byte(`</a><br>
			<span class="rowsmall lastReplyAt">`)
topics_frags[90] = []byte(`</span>
		</span>
	</div>
	</div>`)
topics_frags[91] = []byte(`<div class="rowitem passive rowmsg">`)
topics_frags[92] = []byte(` <a href="/topics/create/">`)
topics_frags[93] = []byte(`</a>`)
topics_frags[94] = []byte(`</div>`)
topics_frags[95] = []byte(`
</div>

`)
paginator_frags[0] = []byte(`<div class="pageset">
	`)
paginator_frags[1] = []byte(`<div class="pageitem"><a href="?page=`)
paginator_frags[2] = []byte(`" rel="prev" aria-label="`)
paginator_frags[3] = []byte(`">`)
paginator_frags[4] = []byte(`</a></div>
	<link rel="prev" href="?page=`)
paginator_frags[5] = []byte(`" />`)
paginator_frags[6] = []byte(`
	<div class="pageitem"><a href="?page=`)
paginator_frags[7] = []byte(`">`)
paginator_frags[8] = []byte(`</a></div>
	`)
paginator_frags[9] = []byte(`
	<link rel="next" href="?page=`)
paginator_frags[10] = []byte(`" />
	<div class="pageitem"><a href="?page=`)
paginator_frags[11] = []byte(`" rel="next" aria-label="`)
paginator_frags[12] = []byte(`">`)
paginator_frags[13] = []byte(`</a></div>`)
paginator_frags[14] = []byte(`
</div>`)
topics_frags[96] = []byte(`

</main>
`)
forum_frags[0] = []byte(`<div id="prevFloat" class="prev_button"><a class="prev_link" aria-label="`)
forum_frags[1] = []byte(`" rel="prev" href="/forum/`)
forum_frags[2] = []byte(`?page=`)
forum_frags[3] = []byte(`">`)
forum_frags[4] = []byte(`</a></div>`)
forum_frags[5] = []byte(`<div id="nextFloat" class="next_button"><a class="next_link" aria-label="`)
forum_frags[6] = []byte(`" rel="next" href="/forum/`)
forum_frags[7] = []byte(`?page=`)
forum_frags[8] = []byte(`">`)
forum_frags[9] = []byte(`</a></div>`)
forum_frags[10] = []byte(`

<main id="forumItemList" itemscope itemtype="http://schema.org/ItemList">
	<div id="forum_head_block" class="rowblock rowhead topic_list_title_block`)
forum_frags[11] = []byte(` has_opt`)
forum_frags[12] = []byte(`">
		<div class="rowitem forum_title">
			<h1 itemprop="name">`)
forum_frags[13] = []byte(`</h1>
		</div>
		`)
forum_frags[14] = []byte(`
			<div class="optbox">
				`)
forum_frags[15] = []byte(`
				<div class="pre_opt auto_hide"></div>
				<div class="opt create_topic_opt" title="`)
forum_frags[16] = []byte(`" aria-label="`)
forum_frags[17] = []byte(`"><a class="create_topic_link" href="/topics/create/`)
forum_frags[18] = []byte(`"></a></div>
				`)
forum_frags[19] = []byte(`
				<div class="opt mod_opt" title="`)
forum_frags[20] = []byte(`">
					<a class="moderate_link" href="#" aria-label="`)
forum_frags[21] = []byte(`"></a>
				</div>
				`)
forum_frags[22] = []byte(`<div class="opt locked_opt" title="`)
forum_frags[23] = []byte(`" aria-label="`)
forum_frags[24] = []byte(`"><a></a></div>`)
forum_frags[25] = []byte(`
			</div>
			<div style="clear: both;"></div>
		`)
forum_frags[26] = []byte(`
	</div>
	`)
forum_frags[27] = []byte(`
	<div class="mod_floater auto_hide">
		<form method="post">
			<div class="mod_floater_head">
				<span>`)
forum_frags[28] = []byte(`</span>
			</div>
			<div class="mod_floater_body">
				<select class="mod_floater_options">
					<option val="delete">`)
forum_frags[29] = []byte(`</option>
					<option val="lock">`)
forum_frags[30] = []byte(`</option>
					<option val="move">`)
forum_frags[31] = []byte(`</option>
				</select>
				<button>`)
forum_frags[32] = []byte(`</button>
			</div>
		</form>
	</div>
	`)
forum_frags[33] = []byte(`
	<div id="forum_topic_create_form" class="rowblock topic_create_form quick_create_form" style="display: none;" aria-label="`)
forum_frags[34] = []byte(`">
		<form id="quick_post_form" enctype="multipart/form-data" action="/topic/create/submit/" method="post"></form>
		<img class="little_row_avatar" src="`)
forum_frags[35] = []byte(`" height="64" alt="`)
forum_frags[36] = []byte(`" title="`)
forum_frags[37] = []byte(`" />
		<input form="quick_post_form" id="topic_board_input" name="topic-board" value="`)
forum_frags[38] = []byte(`" type="hidden">
		<div class="main_form">
			<div class="topic_meta">
				<div class="formrow topic_name_row real_first_child">
					<div class="formitem">
						<input form="quick_post_form" name="topic-name" placeholder="`)
forum_frags[39] = []byte(`" required>
					</div>
				</div>
			</div>
			<div class="formrow topic_content_row">
				<div class="formitem">
					<textarea form="quick_post_form" id="input_content" name="topic-content" placeholder="`)
forum_frags[40] = []byte(`" required></textarea>
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
forum_frags[41] = []byte(`</button>
					<button form="quick_post_form" class="formbutton" id="add_poll_button">`)
forum_frags[42] = []byte(`</button>
					`)
forum_frags[43] = []byte(`
					<input name="upload_files" form="quick_post_form" id="upload_files" multiple type="file" style="display: none;" />
					<label for="upload_files" class="formbutton add_file_button">`)
forum_frags[44] = []byte(`</label>
					<div id="upload_file_dock"></div>`)
forum_frags[45] = []byte(`
					<button class="formbutton close_form">`)
forum_frags[46] = []byte(`</button>
				</div>
			</div>
		</div>
	</div>
	`)
forum_frags[47] = []byte(`
	<div id="forum_topic_list" class="rowblock topic_list" aria-label="`)
forum_frags[48] = []byte(`">
		`)
forum_frags[49] = []byte(`<div class="topic_row" data-tid="`)
forum_frags[50] = []byte(`">
		<div class="rowitem topic_left passive datarow `)
forum_frags[51] = []byte(`topic_sticky`)
forum_frags[52] = []byte(`topic_closed`)
forum_frags[53] = []byte(`">
			<span class="selector"></span>
			<a href="`)
forum_frags[54] = []byte(`"><img src="`)
forum_frags[55] = []byte(`" height="64" alt="`)
forum_frags[56] = []byte(`'s Avatar" title="`)
forum_frags[57] = []byte(`'s Avatar" /></a>
			<span class="topic_inner_left">
				<a class="rowtopic" href="`)
forum_frags[58] = []byte(`" itemprop="itemListElement" title="`)
forum_frags[59] = []byte(`"><span>`)
forum_frags[60] = []byte(`</span></a>
				<br /><a class="rowsmall starter" href="`)
forum_frags[61] = []byte(`" title="`)
forum_frags[62] = []byte(`">`)
forum_frags[63] = []byte(`</a>
				`)
forum_frags[64] = []byte(`<span class="rowsmall topic_status_e topic_status_closed" title="`)
forum_frags[65] = []byte(`"> | &#x1F512;&#xFE0E</span>`)
forum_frags[66] = []byte(`<span class="rowsmall topic_status_e topic_status_sticky" title="`)
forum_frags[67] = []byte(`"> | &#x1F4CD;&#xFE0E</span>`)
forum_frags[68] = []byte(`
			</span>
			<span class="topic_inner_right rowsmall" style="float: right;">
				<span class="replyCount">`)
forum_frags[69] = []byte(`</span><br />
				<span class="likeCount">`)
forum_frags[70] = []byte(`</span>
			</span>
		</div>
		<div class="rowitem topic_right passive datarow `)
forum_frags[71] = []byte(`topic_sticky`)
forum_frags[72] = []byte(`topic_closed`)
forum_frags[73] = []byte(`">
			<a href="`)
forum_frags[74] = []byte(`"><img src="`)
forum_frags[75] = []byte(`" height="64" alt="`)
forum_frags[76] = []byte(`'s Avatar" title="`)
forum_frags[77] = []byte(`'s Avatar" /></a>
			<span>
				<a href="`)
forum_frags[78] = []byte(`" class="lastName" style="font-size: 14px;" title="`)
forum_frags[79] = []byte(`">`)
forum_frags[80] = []byte(`</a><br>
				<span class="rowsmall lastReplyAt">`)
forum_frags[81] = []byte(`</span>
			</span>
		</div>
		</div>`)
forum_frags[82] = []byte(`<div class="rowitem passive rowmsg">`)
forum_frags[83] = []byte(` <a href="/topics/create/`)
forum_frags[84] = []byte(`">`)
forum_frags[85] = []byte(`</a>`)
forum_frags[86] = []byte(`</div>`)
forum_frags[87] = []byte(`
	</div>

`)
forum_frags[88] = []byte(`

</main>
`)
login_frags[0] = []byte(`
<main id="login_page">
	<div class="rowblock rowhead">
		<div class="rowitem"><h1>`)
login_frags[1] = []byte(`</h1></div>
	</div>
	<div class="rowblock">
		<form action="/accounts/login/submit/" method="post">
			<div class="formrow login_name_row">
				<div class="formitem formlabel"><a id="login_name_label">`)
login_frags[2] = []byte(`</a></div>
				<div class="formitem"><input name="username" type="text" placeholder="`)
login_frags[3] = []byte(`" aria-labelledby="login_name_label" required /></div>
			</div>
			<div class="formrow login_password_row">
				<div class="formitem formlabel"><a id="login_password_label">`)
login_frags[4] = []byte(`</a></div>
				<div class="formitem"><input name="password" type="password" autocomplete="current-password" placeholder="*****" aria-labelledby="login_password_label" required /></div>
			</div>
			<div class="formrow login_button_row">
				<div class="formitem"><button name="login-button" class="formbutton">`)
login_frags[5] = []byte(`</button></div>
				<div class="formitem dont_have_account">`)
login_frags[6] = []byte(`</div>
			</div>
		</form>
	</div>
</main>
`)
register_frags[0] = []byte(`
<main id="register_page">
	<div class="rowblock rowhead">
		<div class="rowitem"><h1>`)
register_frags[1] = []byte(`</h1></div>
	</div>
	<div class="rowblock">
		<form action="/accounts/create/submit/" method="post">
			<div class="formrow">
				<div class="formitem formlabel"><a id="username_label">`)
register_frags[2] = []byte(`</a></div>
				<div class="formitem"><input name="username" type="text" placeholder="`)
register_frags[3] = []byte(`" aria-labelledby="username_label" required /></div>
			</div>
			<div class="formrow">
				<div class="formitem formlabel"><a id="email_label">`)
register_frags[4] = []byte(`</a></div>
				<div class="formitem"><input name="email" type="email" placeholder="joe.doe@example.com" aria-labelledby="email_label" required /></div>
			</div>
			<div class="formrow">
				<div class="formitem formlabel"><a id="password_label">`)
register_frags[5] = []byte(`</a></div>
				<div class="formitem"><input name="password" type="password" autocomplete="new-password" placeholder="*****" aria-labelledby="password_label" required /></div>
			</div>
			<div class="formrow">
				<div class="formitem formlabel"><a id="confirm_password_label">`)
register_frags[6] = []byte(`</a></div>
				<div class="formitem"><input name="confirm_password" type="password" placeholder="*****" aria-labelledby="confirm_password_label" required /></div>
			</div>
			<div class="formrow">
				<div class="formitem"><button name="register-button" class="formbutton">`)
register_frags[7] = []byte(`</button></div>
			</div>
		</form>
	</div>
</main>
`)
error_frags[0] = []byte(`
<main>
	<div class="rowblock rowhead">
		<div class="rowitem"><h1>`)
error_frags[1] = []byte(`</h1></div>
	</div>
	<div class="rowblock">
		<div class="rowitem passive rowmsg">`)
error_frags[2] = []byte(`</div>
	</div>
</main>
`)
ip_search_frags[0] = []byte(`
<main id="ip_search_container">
	<div class="rowblock rowhead">
		<div class="rowitem">
			<h1>`)
ip_search_frags[1] = []byte(`</h1>
		</div>
	</div>
	<form action="/users/ips/" method="get" id="ip-search-form"></form>
	<div class="rowblock ip_search_block">
		<div class="rowitem passive">
			<input form="ip-search-form" name="ip" class="ip_search_input" type="search" placeholder="🔍︎"`)
ip_search_frags[2] = []byte(` value="`)
ip_search_frags[3] = []byte(`"`)
ip_search_frags[4] = []byte(` />
			<input form="ip-search-form" class="ip_search_search" type="submit" value="`)
ip_search_frags[5] = []byte(`" />
		</div>
	</div>
	`)
ip_search_frags[6] = []byte(`
	<div class="rowblock rowlist bgavatars">
		`)
ip_search_frags[7] = []byte(`<div class="rowitem" style="background-image: url('`)
ip_search_frags[8] = []byte(`');">
			<img src="`)
ip_search_frags[9] = []byte(`" class="bgsub" alt="`)
ip_search_frags[10] = []byte(`'s Avatar" />
			<a class="rowTitle" href="`)
ip_search_frags[11] = []byte(`">`)
ip_search_frags[12] = []byte(`</a>
		</div>
		`)
ip_search_frags[13] = []byte(`<div class="rowitem rowmsg">`)
ip_search_frags[14] = []byte(`</div>`)
ip_search_frags[15] = []byte(`
	</div>
	`)
ip_search_frags[16] = []byte(`
</main>
`)
guilds_guild_list_frags[0] = []byte(`
<main>
	<div class="rowblock opthead">
		<div class="rowitem"><a>Guild List</a></div>
	</div>
	<div class="rowblock">
		`)
guilds_guild_list_frags[1] = []byte(`<div class="rowitem datarow">
			<span style="float: left;">
				<a href="`)
guilds_guild_list_frags[2] = []byte(`" style="">`)
guilds_guild_list_frags[3] = []byte(`</a>
				<br /><span class="rowsmall">`)
guilds_guild_list_frags[4] = []byte(`</span>
			</span>
			<span style="float: right;">
				<span style="float: right;font-size: 14px;">`)
guilds_guild_list_frags[5] = []byte(` members</span>
				<br /><span class="rowsmall">`)
guilds_guild_list_frags[6] = []byte(`</span>
			</span>
			<div style="clear: both;"></div>
		</div>
		`)
guilds_guild_list_frags[7] = []byte(`<div class="rowitem passive">There aren't any visible guilds.</div>`)
guilds_guild_list_frags[8] = []byte(`
	</div>
</main>
`)
}
