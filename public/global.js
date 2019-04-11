'use strict';
var formVars = {};
var alertMapping = {};
var alertList = [];
var alertCount = 0;
var moreTopicCount = 0;
var conn = false;
var selectedTopics = [];
var attachItemCallback = function(){}
var quoteItemCallback = function(){}
var baseTitle = document.title;
var wsBackoff = 0;

// Topic move
var forumToMoveTo = 0;

// TODO: Write a friendlier error handler which uses a .notice or something, we could have a specialised one for alerts
function ajaxError(xhr,status,errstr) {
	console.log("The AJAX request failed");
	console.log("xhr", xhr);
	console.log("status", status);
	console.log("errstr", errstr);
	if(status=="parsererror") console.log("The server didn't respond with a valid JSON response");
	console.trace();
}

function postLink(event) {
	event.preventDefault();
	let formAction = $(event.target).closest('a').attr("href");
	$.ajax({ url: formAction, type: "POST", dataType: "json", error: ajaxError, data: {js: "1"} });
}

function bindToAlerts() {
	console.log("bindToAlerts");
	$(".alertItem.withAvatar a").unbind("click");
	$(".alertItem.withAvatar a").click(function(event) {
		event.stopPropagation();
		event.preventDefault();
		$.ajax({
			url: "/api/?action=set&module=dismiss-alert",
			type: "POST",
			dataType: "json",
			data: { asid: $(this).attr("data-asid") },
			error: ajaxError,
			success: () => {
				window.location.href = this.getAttribute("href");
			}
		});
	});
}

function addAlert(msg, notice = false) {
	var mmsg = msg.msg;
	if("sub" in msg) {
		for(var i = 0; i < msg.sub.length; i++) mmsg = mmsg.replace("\{"+i+"\}", msg.sub[i]);
	}

	let aItem = Template_alert({
		ASID: msg.asid,
		Path: msg.path,
		Avatar: msg.avatar || "",
		Message: mmsg
	})
	//alertMapping[msg.asid] = aItem;
	let div = document.createElement('div');
	div.innerHTML = aItem.trim();
	alertMapping[msg.asid] = div.firstChild;
	alertList.push(msg.asid);

	if(notice) {
		// TODO: Add some sort of notification queue to avoid flooding the end-user with notices?
		// TODO: Use the site name instead of "Something Happened"
		if(Notification.permission === "granted") {
			var n = new Notification("Something Happened",{
				body: mmsg,
				icon: msg.avatar,
			});
			setTimeout(n.close.bind(n), 8000);
		}
	}

	runInitHook("after_add_alert");
}

function updateAlertList(menuAlerts) {
	let alertListNode = menuAlerts.getElementsByClassName("alertList")[0];
	let alertCounterNode = menuAlerts.getElementsByClassName("alert_counter")[0];
	alertCounterNode.textContent = "0";
	
	alertListNode.innerHTML = "";
	let any = false;
	/*let outList = "";
	let j = 0;
	for(var i = 0; i < alertList.length && j < 8; i++) {
		outList += alertMapping[alertList[i]];
		j++;
	}*/
	let j = 0;
	for(var i = 0; i < alertList.length && j < 8; i++) {
		any = true;
		alertListNode.appendChild(alertMapping[alertList[i]]);
		//outList += alertMapping[alertList[i]];
		j++;
	}
	if(!any) alertListNode.innerHTML = "<div class='alertItem'>"+phraseBox["alerts"]["alerts.no_alerts"]+"</div>";

	if(alertCount != 0) {
		alertCounterNode.textContent = alertCount;
		menuAlerts.classList.add("has_alerts");
		let nTitle = "("+alertCount+") "+baseTitle;
		if(document.title!=nTitle) document.title = nTitle;
	} else {
		menuAlerts.classList.remove("has_alerts");
		if(document.title!=baseTitle) document.title = baseTitle;
	}

	bindToAlerts();
	console.log("alertCount:",alertCount)
	runInitHook("after_update_alert_list", alertCount);
}

function setAlertError(menuAlerts,msg) {
	let alertListNode = menuAlerts.getElementsByClassName("alertList")[0];
	alertListNode.innerHTML = "<div class='alertItem'>"+msg+"</div>";
}

var alertsInitted = false;
function loadAlerts(menuAlerts) {
	if(!alertsInitted) return;
	$.ajax({
		type: 'get',
		dataType: 'json',
		url:'/api/?action=get&module=alerts',
		success: (data) => {
			if("errmsg" in data) {
				setAlertError(menuAlerts,data.errmsg)
				return;
			}
			alertList = [];
			alertMapping = {};
			for(var i in data.msgs) addAlert(data.msgs[i]);
			console.log("data.msgCount:",data.msgCount)
			alertCount = data.msgCount;
			updateAlertList(menuAlerts)
		},
		error: (magic,theStatus,error) => {
			let errtxt = "Unable to get the alerts";
			try {
				var data = JSON.parse(magic.responseText);
				if("errmsg" in data) errtxt = data.errmsg;
			} catch(err) {
				console.log(magic.responseText);
				console.log(err);
			}
			console.log("error", error);
			setAlertError(menuAlerts,errtxt);
		}
	});
}

function SplitN(data,ch,n) {
	var out = [];
	if(data.length === 0) return out;

	var lastIndex = 0;
	var j = 0;
	var lastN = 1;
	for(let i = 0; i < data.length; i++) {
		if(data[i] === ch) {
			out[j++] = data.substring(lastIndex,i);
			lastIndex = i;
			if(lastN === n) break;
			lastN++;
		}
	}
	if(data.length > lastIndex) out[out.length - 1] += data.substring(lastIndex);
	return out;
}

function wsAlertEvent(data) {
	console.log("wsAlertEvent:",data)
	addAlert(data, true);
	alertCount++;

	let aTmp = alertList;
	alertList = [alertList[alertList.length-1]];
	aTmp = aTmp.slice(0,-1);
	for(let i = 0; i < aTmp.length; i++) alertList.push(aTmp[i]);
	//var alist = "";
	//for (var i = 0; i < alertList.length; i++) alist += alertMapping[alertList[i]];
	// TODO: Add support for other alert feeds like PM Alerts
	var generalAlerts = document.getElementById("general_alerts");
	// TODO: Make sure we update alertCount here
	updateAlertList(generalAlerts/*, alist*/);
}

function runWebSockets(resume = false) {
	if(window.location.protocol == "https:") {
		conn = new WebSocket("wss://" + document.location.host + "/ws/");
	} else conn = new WebSocket("ws://" + document.location.host + "/ws/");

	conn.onerror = (err) => {
		console.log(err);
	}

	// TODO: Sync alerts, topic list, etc.
	conn.onopen = () => {
		console.log("The WebSockets connection was opened");
		conn.send("page " + document.location.pathname + '\r');
		if(resume) conn.send("resume " + Math.round((new Date()).getTime() / 1000) + '\r');
		// TODO: Don't ask again, if it's denied. We could have a setting in the UCP which automatically requests this when someone flips desktop notifications on
		if(me.User.ID > 0) Notification.requestPermission();
	}

	conn.onclose = () => {
		conn = false;
		console.log("The WebSockets connection was closed");
		let backoff = 0.8;
		if(wsBackoff < 0) wsBackoff = 0;
		else if(wsBackoff > 12) backoff = 11;
		else if(wsBackoff > 5) backoff = 5;
		wsBackoff++;

		setTimeout(() => {
			var alertMenuList = document.getElementsByClassName("menu_alerts");
			for(var i = 0; i < alertMenuList.length; i++) loadAlerts(alertMenuList[i]);
			runWebSockets(true);
		}, backoff * 60 * 1000);

		if(wsBackoff > 0) {
			if(wsBackoff <= 5) setTimeout(() => wsBackoff--, 5.5 * 60 * 1000);
			else if(wsBackoff <= 12) setTimeout(() => wsBackoff--, 11.5 * 60 * 1000);
			else setTimeout(() => wsBackoff--, 20 * 60 * 1000);
		}
	}

	conn.onmessage = (event) => {
		if(event.data[0] == "{") {
			console.log("json message");
			let data = "";
			try {
				data = JSON.parse(event.data);
			} catch(err) {
				console.log(err);
				return;
			}

			if ("msg" in data) wsAlertEvent(data);
			else if("event" in data) {
				if(data.event == "dismiss-alert"){
					Object.keys(alertMapping).forEach((key) => {
						if(key==data.asid) {
							alertCount--;
							let index = -1;
							for(var i = 0; i < alertList.length; i++) {
								if(alertList[i]==key) {
									alertList[i] = 0;
									index = i;
								}
							}
							if(index==-1) return;

							for(var i = index; (i+1) < alertList.length; i++) {
								alertList[i] = alertList[i+1];
							}
							alertList.splice(alertList.length-1,1);
							delete alertMapping[key];

							// TODO: Add support for other alert feeds like PM Alerts
							var generalAlerts = document.getElementById("general_alerts");
							if(alertList.length < 8) loadAlerts(generalAlerts);
							else updateAlertList(generalAlerts);
						}
					});
				}
			} else if("Topics" in data) {
				console.log("topic in data");
				console.log("data:", data);
				let topic = data.Topics[0];
				if(topic === undefined){
					console.log("empty topic list");
					return;
				}
				// TODO: Fix the data race where the function hasn't been loaded yet
				let renTopic = Template_topics_topic(topic);
				$(".topic_row[data-tid='"+topic.ID+"']").addClass("ajax_topic_dupe");

				let node = $(renTopic);
				node.addClass("new_item hide_ajax_topic");
				console.log("Prepending to topic list");
				$(".topic_list").prepend(node);
				moreTopicCount++;

				let moreTopicBlocks = document.getElementsByClassName("more_topic_block_initial");
				for(let i = 0; i < moreTopicBlocks.length; i++) {
					let moreTopicBlock = moreTopicBlocks[i];
					moreTopicBlock.classList.remove("more_topic_block_initial");
					moreTopicBlock.classList.add("more_topic_block_active");

					console.log("phraseBox:",phraseBox);
					let msgBox = moreTopicBlock.getElementsByClassName("more_topics")[0];
					msgBox.innerText = phraseBox["topic_list"]["topic_list.changed_topics"].replace("%d",moreTopicCount);
				}
			} else {
				console.log("unknown message");
				console.log(data);
			}
		}

		var messages = event.data.split('\r');
		for(var i = 0; i < messages.length; i++) {
			let message = messages[i];
			//console.log("Message: ",message);
			let msgblocks = SplitN(message," ",3);
			if(msgblocks.length < 3) continue;
			if(message.startsWith("set ")) {
				let oldInnerHTML = document.querySelector(msgblocks[1]).innerHTML;
				if(msgblocks[2]==oldInnerHTML) continue;
				document.querySelector(msgblocks[1]).innerHTML = msgblocks[2];
			} else if(message.startsWith("set-class ")) {
				// Fix to stop the inspector from getting all jittery
				let oldClassName = document.querySelector(msgblocks[1]).className;
				if(msgblocks[2]==oldClassName) continue;
				document.querySelector(msgblocks[1]).className = msgblocks[2];
			}
		}
	}
}

(() => {
	addInitHook("pre_init", () => {
		console.log("before notify on alert")
		// We can only get away with this because template_alert has no phrases, otherwise it too would have to be part of the "dance", I miss Go concurrency :(
		notifyOnScriptW("template_alert", (e) => {
			if(e!=undefined) console.log("failed alert? why?", e)
		}, () => {
			//console.log("ha")
			if(!Template_alert) throw("template function not found");
			addInitHook("after_phrases", () => {
				// TODO: The load part of loadAlerts could be done asynchronously while the update of the DOM could be deferred
				$(document).ready(() => {
					alertsInitted = true;
					var alertMenuList = document.getElementsByClassName("menu_alerts");
					for(var i = 0; i < alertMenuList.length; i++) loadAlerts(alertMenuList[i]);
					if(window["WebSocket"]) runWebSockets();
				});
			});
		});

		$(document).ready(mainInit);
	});
})();

// TODO: Use these in .filter_item and pass back an item count from the backend to work with here
// Ported from common/parser.go
function PageOffset(count, page, perPage) {
	let offset = 0;
	let lastPage = LastPage(count, perPage)
	if(page > 1) {
		offset = (perPage * page) - perPage
	} else if (page == -1) {
		page = lastPage
		offset = (perPage * page) - perPage
	} else {
		page = 1
	}

	// We don't want the offset to overflow the slices, if everything's in memory
	//if(offset >= (count - 1)) offset = 0;
	return {Offset:offset, Page:page, LastPage:lastPage}
}
function LastPage(count, perPage) {
	return (count / perPage) + 1
}
function Paginate(count, perPage, maxPages) {
	if(count < perPage) return [1];
	let page = 0;
	let out = [];
	for(let current = 0; current < count; current += perPage){
		page++;
		out.push(page);
		if(out.length >= maxPages) break;
	}
	return out;
}

function mainInit(){
	runInitHook("start_init");

	$(".more_topics").click((event) => {
		event.preventDefault();
		let moreTopicBlocks = document.getElementsByClassName("more_topic_block_active");
		for(let i = 0; i < moreTopicBlocks.length; i++) {
			let moreTopicBlock = moreTopicBlocks[i];
			moreTopicBlock.classList.remove("more_topic_block_active");
			moreTopicBlock.classList.add("more_topic_block_initial");
		}
		$(".ajax_topic_dupe").fadeOut("slow", function(){
			$(this).remove();
		});
		$(".hide_ajax_topic").removeClass("hide_ajax_topic"); // TODO: Do Fade
		moreTopicCount = 0;
	})

	$(".add_like").click(function(event) {
		event.preventDefault();
		let target = this.closest("a").getAttribute("href");
		console.log("target: ", target);
		this.classList.remove("add_like");
		this.classList.add("remove_like");
		let controls = this.closest(".controls");
		let hadLikes = controls.classList.contains("has_likes");
		if(!hadLikes) controls.classList.add("has_likes");
		let likeCountNode = controls.getElementsByClassName("like_count")[0];
		console.log("likeCountNode",likeCountNode);
		likeCountNode.innerHTML = parseInt(likeCountNode.innerHTML) + 1;
		let likeButton = this;
		
		$.ajax({
			url: target,
			type: "POST",
			dataType: "json",
			data: { isJs: 1 },
			error: ajaxError,
			success: function (data, status, xhr) {
				if("success" in data) {
					if(data["success"] == "1") return;
				}
				// addNotice("Failed to add a like: {err}")
				likeButton.classList.add("add_like");
				likeButton.classList.remove("remove_like");
				if(!hadLikes) controls.classList.remove("has_likes");
				likeCountNode.innerHTML = parseInt(likeCountNode.innerHTML) - 1;
				console.log("data", data);
				console.log("status", status);
				console.log("xhr", xhr);
			}
		});
	});

	$(".link_label").click(function(event) {
		event.preventDefault();
		let linkSelect = $('#'+$(this).attr("data-for"));
		if(!linkSelect.hasClass("link_opened")) {
			event.stopPropagation();
			linkSelect.addClass("link_opened");
		}
	});

	function rebuildPaginator(lastPage) {
		let urlParams = new URLSearchParams(window.location.search);
		let page = urlParams.get('page');
		if(page=="") page = 1;
		let stopAtPage = lastPage;
		if(stopAtPage>5) stopAtPage = 5;

		let pageList = [];
		for(let i = 0; i < stopAtPage;i++) pageList.push(i+1);
		//$(".pageset").html(Template_paginator({PageList: pageList, Page: page, LastPage: lastPage}));
		let ok = false;
		$(".pageset").each(function(){
			this.outerHTML = Template_paginator({PageList: pageList, Page: page, LastPage: lastPage});
			ok = true;
		});
		if(!ok) $(Template_paginator({PageList: pageList, Page: page, LastPage: lastPage})).insertAfter("#topic_list");
	}

	function rebindPaginator() {
		$(".pageitem a").unbind("click");
		$(".pageitem a").click(function() {
			event.preventDefault();
			// TODO: Take mostviewed into account
			let url = "//"+window.location.host+window.location.pathname;
			let urlParams = new URLSearchParams(window.location.search);
			urlParams.set("page",new URLSearchParams(this.getAttribute("href")).get("page"));
			let q = "?";
			for(let item of urlParams.entries()) q += item[0]+"="+item[1]+"&";
			if(q.length>1) q = q.slice(0,-1);

			// TODO: Try to de-duplicate some of these fetch calls
			fetch(url+q+"&js=1", {credentials: "same-origin"})
				.then((resp) => {
					if(!resp.ok) throw(url+q+"&js=1 failed to load");
					return resp.json();
				}).then((data) => {
					if(!"Topics" in data) throw("no Topics in data");
					let topics = data["Topics"];
					console.log("ajax navigated to different page");

					// TODO: Fix the data race where the function hasn't been loaded yet
					let out = "";
					for(let i = 0; i < topics.length;i++) out += Template_topics_topic(topics[i]);
					$(".topic_list").html(out);

					let obj = {Title: document.title, Url: url+q};
					history.pushState(obj, obj.Title, obj.Url);
					rebuildPaginator(data.LastPage);
					rebindPaginator();
				}).catch((ex) => {
					console.log("Unable to get script '"+url+q+"&js=1"+"'");
					console.log("ex: ", ex);
					console.trace();
				});
		});
	}

	// TODO: Render a headless topics.html instead of the individual topic rows and a bit of JS glue
	$(".filter_item").click(function(event) {
		if(!window.location.pathname.startsWith("/topics/")) return
		event.preventDefault();
		let that = this;
		let fid = this.getAttribute("data-fid");
		// TODO: Take mostviewed into account
		let url = "//"+window.location.host+"/topics/?fids="+fid;

		fetch(url+"&js=1", {credentials: "same-origin"})
		.then((resp) => {
			if(!resp.ok) throw(url+"&js=1 failed to load");
			return resp.json();
		}).then((data) => {
			console.log("data:",data);
			if(!"Topics" in data) throw("no Topics in data");
			let topics = data["Topics"];
			console.log("ajax navigated to "+that.innerText);
			
			// TODO: Fix the data race where the function hasn't been loaded yet
			let out = "";
			for(let i = 0; i < topics.length;i++) out += Template_topics_topic(topics[i]);
			$(".topic_list").html(out);
			//$(".topic_list").addClass("single_forum");

			baseTitle = that.innerText;
			if(alertCount > 0) document.title = "("+alertCount+") "+baseTitle;
			else document.title = baseTitle;
			let obj = {Title: document.title, Url: url};
			history.pushState(obj, obj.Title, obj.Url);
			rebuildPaginator(data.LastPage)
			rebindPaginator();

			$(".filter_item").each(function(){
				this.classList.remove("filter_selected");
			});
			that.classList.add("filter_selected");
			$(".topic_list_title h1").text(that.innerText);
		}).catch((ex) => {
			console.log("Unable to get script '"+url+"&js=1"+"'");
			console.log("ex: ", ex);
			console.trace();
		});
	});

	if (document.getElementById("topicsItemList")!==null) rebindPaginator();
	if (document.getElementById("forumItemList")!==null) rebindPaginator();

	// TODO: Show a search button when JS is disabled?
	$(".widget_search_input").keypress(function(e) {
		if (e.keyCode != '13') return;
		event.preventDefault();
		// TODO: Take mostviewed into account
		let url = "//"+window.location.host+window.location.pathname;
		let urlParams = new URLSearchParams(window.location.search);
		urlParams.set("q",this.value);
		let q = "?";
		for(let item of urlParams.entries()) q += item[0]+"="+item[1]+"&";
		if(q.length>1) q = q.slice(0,-1);

		// TODO: Try to de-duplicate some of these fetch calls
		fetch(url+q+"&js=1", {credentials: "same-origin"})
			.then((resp) => {
				if(!resp.ok) throw(url+q+"&js=1 failed to load");
				return resp.json();
			}).then((data) => {
				if(!"Topics" in data) throw("no Topics in data");
				let topics = data["Topics"];
				console.log("ajax navigated to search page");

				// TODO: Fix the data race where the function hasn't been loaded yet
				let out = "";
				for(let i = 0; i < topics.length;i++) out += Template_topics_topic(topics[i]);
				$(".topic_list").html(out);

				baseTitle = phraseBox["topic_list"]["topic_list.search_head"];
				$(".topic_list_title h1").text(phraseBox["topic_list"]["topic_list.search_head"]);
				if(alertCount > 0) document.title = "("+alertCount+") "+baseTitle;
				else document.title = baseTitle;
				let obj = {Title: document.title, Url: url+q};
				history.pushState(obj, obj.Title, obj.Url);
				rebuildPaginator(data.LastPage);
				rebindPaginator();
		}).catch((ex) => {
			console.log("Unable to get script '"+url+q+"&js=1"+"'");
			console.log("ex: ", ex);
			console.trace();
		});
	});

	$(".open_edit").click((event) => {
		event.preventDefault();
		$('.hide_on_edit').addClass("edit_opened");
		$('.show_on_edit').addClass("edit_opened");
		runHook("open_edit");
	});

	$(".topic_item .submit_edit").click(function(event){
		event.preventDefault();
		let topicNameInput = $(".topic_name_input").val();
		$(".topic_name").html(topicNameInput);
		$(".topic_name").attr(topicNameInput);
		let topicContentInput = $('.topic_content_input').val();
		$(".topic_content").html(quickParse(topicContentInput));
		let topicStatusInput = $('.topic_status_input').val();
		$(".topic_status_e:not(.open_edit)").html(topicStatusInput);

		$('.hide_on_edit').removeClass("edit_opened");
		$('.show_on_edit').removeClass("edit_opened");
		runHook("close_edit");

		$.ajax({
			url: this.form.getAttribute("action"),
			type: "POST",
			dataType: "json",
			data: {
				topic_name: topicNameInput,
				topic_status: topicStatusInput,
				topic_content: topicContentInput,
				js: 1
			},
			error: ajaxError,
			success: (data,status,xhr) => {
				if("Content" in data) $(".topic_content").html(data["Content"]);
			}
		});
	});

	$(".delete_item").click(function(event) {
		postLink(event);
		$(this).closest('.deletable_block').remove();
	});

	// Miniature implementation of the parser to avoid sending as much data back and forth
	function quickParse(msg) {
		msg = msg.replace(":)", "😀")
		msg = msg.replace(":(", "😞")
		msg = msg.replace(":D", "😃")
		msg = msg.replace(":P", "😛")
		msg = msg.replace(":O", "😲")
		msg = msg.replace(":p", "😛")
		msg = msg.replace(":o", "😲")
		msg = msg.replace(";)", "😉")
		msg = msg.replace("\n","<br>")
		return msg
	}

	$(".edit_item").click(function(event){
		event.preventDefault();
		let blockParent = this.closest('.editable_parent');
		$(blockParent).find('.hide_on_edit').addClass("edit_opened");
		$(blockParent).find('.show_on_edit').addClass("edit_opened");
		let srcNode = blockParent.querySelector(".edit_source");
		let block = blockParent.querySelector('.editable_block');
		block.classList.add("in_edit");
		let source = "";
		if(srcNode!=null) source = srcNode.innerText;
		else source = block.innerHTML;
		// TODO: Add a client template for this
		block.innerHTML = "<textarea style='width: 99%;' name='edit_item'>" + source + "</textarea><br><a href='" + this.closest('a').getAttribute("href") + "'><button class='submit_edit' type='submit'>Update</button></a>";
		runHook("edit_item_pre_bind");

		$(".submit_edit").click(function(event){
			event.preventDefault();
			$(blockParent).find('.hide_on_edit').removeClass("edit_opened");
			$(blockParent).find('.show_on_edit').removeClass("edit_opened");
			block.classList.remove("in_edit");
			let newContent = block.querySelector('textarea').value;
			block.innerHTML = quickParse(newContent);
			if(srcNode!=null) srcNode.innerText = newContent;

			let formAction = this.closest('a').getAttribute("href");
			// TODO: Bounce the parsed post back and set innerHTML to it?
			$.ajax({
				url: formAction,
				type: "POST",
				dataType: "json",
				data: { js: "1", edit_item: newContent },
				error: ajaxError,
				success: (data,status,xhr) => {
					if("Content" in data) block.innerHTML = data["Content"];
				}
			});
		});
	});

	$(".edit_field").click(function(event) {
		event.preventDefault();
		let blockParent = $(this).closest('.editable_parent');
		let block = blockParent.find('.editable_block').eq(0);
		block.html("<input name='edit_field' value='" + block.text() + "' type='text'/><a href='" + $(this).closest('a').attr("href") + "'><button class='submit_edit' type='submit'>Update</button></a>");

		$(".submit_edit").click(function(event) {
			event.preventDefault();
			let blockParent = $(this).closest('.editable_parent');
			let block = blockParent.find('.editable_block').eq(0);
			let newContent = block.find('input').eq(0).val();
			block.html(newContent);

			let formAction = $(this).closest('a').attr("href");
			$.ajax({
				url: formAction + "?session=" + me.User.Session,
				type: "POST",
				dataType: "json",
				error: ajaxError,
				data: { isJs: "1", edit_item: newContent }
			});
		});
	});

	$(".edit_fields").click(function(event)
	{
		event.preventDefault();
		if($(this).find("input").length !== 0) return;
		//console.log("clicked .edit_fields");
		var blockParent = $(this).closest('.editable_parent');
		blockParent.find('.hide_on_edit').addClass("edit_opened");
		blockParent.find('.show_on_edit').addClass("edit_opened");
		blockParent.find('.editable_block').show();
		blockParent.find('.editable_block').each(function(){
			var fieldName = this.getAttribute("data-field");
			var fieldType = this.getAttribute("data-type");
			if(fieldType=="list") {
				var fieldValue = this.getAttribute("data-value");
				if(fieldName in formVars) var it = formVars[fieldName];
				else var it = ['No','Yes'];
				var itLen = it.length;
				var out = "";
				for (var i = 0; i < itLen; i++) {
					var sel = "";
					if(fieldValue == i || fieldValue == it[i]) {
						sel = "selected ";
						this.classList.remove(fieldName + '_' + it[i]);
						this.innerHTML = "";
					}
					out += "<option "+sel+"value='"+i+"'>"+it[i]+"</option>";
				}
				this.innerHTML = "<select data-field='"+fieldName+"' name='"+fieldName+"'>"+out+"</select>";
			}
			else if(fieldType=="hidden") {}
			else this.innerHTML = "<input name='"+fieldName+"' value='"+this.textContent+"' type='text'/>";
		});

		// Remove any handlers already attached to the submitter
		$(".submit_edit").unbind("click");

		$(".submit_edit").click(function(event) {
			event.preventDefault();
			var outData = {isJs: "1"}
			var blockParent = $(this).closest('.editable_parent');
			blockParent.find('.editable_block').each(function() {
				var fieldName = this.getAttribute("data-field");
				var fieldType = this.getAttribute("data-type");
				if(fieldType=="list") {
					var newContent = $(this).find('select :selected').text();
					this.classList.add(fieldName + '_' + newContent);
					this.innerHTML = "";
				} else if(fieldType=="hidden") {
					var newContent = $(this).val();
				} else {
					var newContent = $(this).find('input').eq(0).val();
					this.innerHTML = newContent;
				}
				this.setAttribute("data-value",newContent);
				outData[fieldName] = newContent;
			});

			var formAction = $(this).closest('a').attr("href");
			//console.log("Form Action:", formAction);
			//console.log(outData);
			$.ajax({ url: formAction + "?session=" + me.User.Session, type:"POST", dataType:"json", data: outData, error: ajaxError });
			blockParent.find('.hide_on_edit').removeClass("edit_opened");
			blockParent.find('.show_on_edit').removeClass("edit_opened");
		});
	});

	// This one's for Tempra Conflux
	// TODO: We might want to use pure JS here
	$(".ip_item").each(function(){
		var ip = this.textContent;
		if(ip.length > 10){
			this.innerHTML = "Show IP";
			this.onclick = function(event) {
				event.preventDefault();
				this.textContent = ip;
			};
		}
	});

	$(".quote_item").click(function(){
		event.preventDefault();
		event.stopPropagation();
		let source = this.closest(".post_item").getElementsByClassName("edit_source")[0];
		let content = document.getElementById("input_content")
		console.log("content.value", content.value);

		let item;
		if(content.value == "") item = "<blockquote>" + source.innerHTML + "</blockquote>"
		else item = "\r\n<blockquote>" + source.innerHTML + "</blockquote>";
		content.value = content.value + item;
		console.log("content.value", content.value);

		// For custom / third party text editors
		quoteItemCallback(source.innerHTML,item);
	});

	$(this).click(() => {
		$(".selectedAlert").removeClass("selectedAlert");
		$("#back").removeClass("alertActive");
		$(".link_select").removeClass("link_opened");
	});

	$(".alert_bell").click(function(){
		var menuAlerts = $(this).parent();
		if(menuAlerts.hasClass("selectedAlert")) {
			event.stopPropagation();
			menuAlerts.removeClass("selectedAlert");
			$("#back").removeClass("alertActive");
		}
	});
	$(".menu_alerts").click(function(event) {
		event.stopPropagation();
		if($(this).hasClass("selectedAlert")) return;
		if(!conn) loadAlerts(this);
		this.className += " selectedAlert";
		document.getElementById("back").className += " alertActive"
	});
	$(".link_select").click(event => event.stopPropagation());

	$("input,textarea,select,option").keyup(event => event.stopPropagation())

	$(".create_topic_link").click((event) => {
		event.preventDefault();
		$(".topic_create_form").show();
	});
	$(".topic_create_form .close_form").click((event) => {
		event.preventDefault();
		$(".topic_create_form").hide();
	});

	function uploadFileHandler(fileList,maxFiles = 5, step1 = () => {}, step2 = () => {}) {
		let files = [];
		for(var i = 0; i < fileList.length && i < 5; i++) {
			files[i] = fileList[i];
		}

		let totalSize = 0;
		for(let i = 0; i < files.length; i++) {
			console.log("files[" + i + "]",files[i]);
			totalSize += files[i]["size"];
		}
		if(totalSize > me.Site.MaxRequestSize) {
			throw("You can't upload this much at once, max: " + me.Site.MaxRequestSize);
		}

		for(let i = 0; i < files.length; i++) {
			let reader = new FileReader();
			reader.onload = (e) => {
				let filename = files[i]["name"];
				step1(e,filename)

				let reader = new FileReader();
				reader.onload = (e2) => {
					crypto.subtle.digest('SHA-256',e2.target.result)
						.then((hash) => {
							const hashArray = Array.from(new Uint8Array(hash))
							return hashArray.map(b => ('00' + b.toString(16)).slice(-2)).join('')
						}).then(hash => step2(e,hash,filename));
				}
				reader.readAsArrayBuffer(files[i]);
			}
			reader.readAsDataURL(files[i]);
		}
	}

	// TODO: Surely, there's a prettier and more elegant way of doing this?
	function getExt(filename) {
		if(!filename.indexOf('.' > -1)) {
			throw("This file doesn't have an extension");
		}
		return filename.split('.').pop();
	}

	// Attachment Manager
	function uploadAttachHandler2() {
		let fileDock = this.closest(".attach_edit_bay");
		try {
			uploadFileHandler(this.files, 5, () => {},
			(e, hash, filename) => {
				console.log("hash",hash);
				
				let formData = new FormData();
				formData.append("session",me.User.Session);
				for(let i = 0; i < this.files.length; i++) {
					formData.append("upload_files",this.files[i]);
				}

				let req = new XMLHttpRequest();
				req.addEventListener("load", () => {
					let data = JSON.parse(req.responseText);
					let fileItem = document.createElement("div");
					let ext = getExt(filename);
					// TODO: Check if this is actually an image, maybe push ImageFileExts to the client from the server in some sort of gen.js?
					// TODO: Use client templates here
					fileItem.className = "attach_item attach_image_holder";
					fileItem.innerHTML = "<img src='"+e.target.result+"' height=24 width=24 /><span class='attach_item_path' aid='"+data[hash+"."+ext]+"' fullpath='//" + window.location.host + "/attachs/" + hash + "." + ext+"'>"+hash+"."+ext+"</span><button class='attach_item_select'>Select</button><button class='attach_item_copy'>Copy</button>";
					fileDock.insertBefore(fileItem,fileDock.querySelector(".attach_item_buttons"));
					
					$(".attach_item_select").unbind("click");
					$(".attach_item_copy").unbind("click");
					bindAttachItems()
				});
				req.open("POST","//"+window.location.host+"/"+fileDock.getAttribute("type")+"/attach/add/submit/"+fileDock.getAttribute("id"));
				req.send(formData);
			});
		} catch(e) {
			// TODO: Use a notice instead
			alert(e);
		}
	}

	// Quick Topic / Quick Reply
	function uploadAttachHandler() {
		try {
			uploadFileHandler(this.files,5,(e,filename) => {
				// TODO: Use client templates here
				let fileDock = document.getElementById("upload_file_dock");
				let fileItem = document.createElement("label");
				console.log("fileItem",fileItem);

				let ext = getExt(filename)
				fileItem.innerText = "." + ext;
				fileItem.className = "formbutton uploadItem";
				// TODO: Check if this is actually an image
				fileItem.style.backgroundImage = "url("+e.target.result+")";

				fileDock.appendChild(fileItem);
			},(e,hash, filename) => {
				console.log("hash",hash);
				let ext = getExt(filename)
				let content = document.getElementById("input_content")
				console.log("content.value", content.value);
				
				let attachItem;
				if(content.value == "") attachItem = "//" + window.location.host + "/attachs/" + hash + "." + ext;
				else attachItem = "\r\n//" + window.location.host + "/attachs/" + hash + "." + ext;
				content.value = content.value + attachItem;
				console.log("content.value", content.value);
				
				// For custom / third party text editors
				attachItemCallback(attachItem);
			});
		} catch(e) {
			// TODO: Use a notice instead
			alert(e);
		}
	}

	let uploadFiles = document.getElementById("upload_files");
	if(uploadFiles != null) {
		uploadFiles.addEventListener("change", uploadAttachHandler, false);
	}
	let uploadFilesOp = document.getElementById("upload_files_op");
	if(uploadFilesOp != null) {
		uploadFilesOp.addEventListener("change", uploadAttachHandler2, false);
	}
	let uploadFilesPost = document.getElementsByClassName("upload_files_post");
	if(uploadFilesPost != null) {
		for(let i = 0; i < uploadFilesPost.length; i++) {
			uploadFilesPost[i].addEventListener("change", uploadAttachHandler2, false);
		}
	}

	function copyToClipboard(str) {
		const el = document.createElement('textarea');
		el.value = str;
		el.setAttribute('readonly', '');
		el.style.position = 'absolute';
		el.style.left = '-9999px';
		document.body.appendChild(el);
		el.select();
		document.execCommand('copy');
		document.body.removeChild(el);
	}

	function bindAttachItems() {
		$(".attach_item_select").click(function(){
			let hold = $(this).closest(".attach_item");
			if(hold.hasClass("attach_item_selected")) {
				hold.removeClass("attach_item_selected");
			} else {
				hold.addClass("attach_item_selected");
			}
		});
		$(".attach_item_copy").click(function(){
			let hold = $(this).closest(".attach_item");
			let pathNode = hold.find(".attach_item_path");
			copyToClipboard(pathNode.attr("fullPath"));
		});
	}
	bindAttachItems();

	$(".attach_item_delete").click(function(){
		let formData = new URLSearchParams();
		formData.append("session",me.User.Session);

		let aidList = "";
		let elems = document.getElementsByClassName("attach_item_selected");
		if(elems == null) return;
		
		for(let i = 0; i < elems.length; i++) {
			let pathNode = elems[i].querySelector(".attach_item_path");
			console.log("pathNode",pathNode);
			aidList += pathNode.getAttribute("aid") + ",";
			elems[i].remove();
		}
		if(aidList.length > 0) aidList = aidList.slice(0, -1);
		console.log("aidList",aidList)
		formData.append("aids",aidList);
		
		let req = new XMLHttpRequest();
		let fileDock = this.closest(".attach_edit_bay");
		req.open("POST","//"+window.location.host+"/"+fileDock.getAttribute("type")+"/attach/remove/submit/"+fileDock.getAttribute("id"),true);
		req.send(formData);
	});
	
	$(".moderate_link").click((event) => {
		event.preventDefault();
		$(".pre_opt").removeClass("auto_hide");
		$(".moderate_link").addClass("moderate_open");
		$(".topic_row").each(function(){
			$(this).click(function(){
				selectedTopics.push(parseInt($(this).attr("data-tid"),10));
				if(selectedTopics.length==1) {
					var msg = phraseBox["topic_list"]["topic_list.what_to_do_single"];
				} else {
					var msg = "What do you want to do with these "+selectedTopics.length+" topics?";
				}
				$(".mod_floater_head span").html(msg);
				$(this).addClass("topic_selected");
				$(".mod_floater").removeClass("auto_hide");
			});
		});

		let bulkActionSender = function(action, selectedTopics, fragBit) {
			let url = "/topic/"+action+"/submit/"+fragBit+"?session=" + me.User.Session;
			$.ajax({
				url: url,
				type: "POST",
				data: JSON.stringify(selectedTopics),
				contentType: "application/json",
				error: ajaxError,
				success: () => {
					window.location.reload();
				}
			});
		};
		$(".mod_floater_submit").click(function(event){
			event.preventDefault();
			let selectNode = this.form.querySelector(".mod_floater_options");
			let optionNode = selectNode.options[selectNode.selectedIndex];
			let action = optionNode.getAttribute("val");

			// Handle these specially
			switch(action) {
				case "move":
					console.log("move action");
					let modTopicMover = $("#mod_topic_mover");
					$("#mod_topic_mover").removeClass("auto_hide");
					$("#mod_topic_mover .pane_row").click(function(){
						modTopicMover.find(".pane_row").removeClass("pane_selected");
						let fid = this.getAttribute("data-fid");
						if (fid == null) return;
						this.classList.add("pane_selected");
						console.log("fid: " + fid);
						forumToMoveTo = fid;

						$("#mover_submit").unbind("click");
						$("#mover_submit").click(function(event){
							event.preventDefault();
							bulkActionSender("move",selectedTopics,forumToMoveTo);
						});
					});
					return;
			}
			
			bulkActionSender(action,selectedTopics,"");
		});
	});

	$("#themeSelectorSelect").change(function(){
		console.log("Changing the theme to " + this.options[this.selectedIndex].getAttribute("val"));
		$.ajax({
			url: this.form.getAttribute("action") + "?session=" + me.User.Session,
			type: "POST",
			dataType: "json",
			data: { "newTheme": this.options[this.selectedIndex].getAttribute("val"), isJs: "1" },
			error: ajaxError,
			success: function (data, status, xhr) {
				console.log("Theme successfully switched");
				console.log("data", data);
				console.log("status", status);
				console.log("xhr", xhr);
				window.location.reload();
			}
		});
	});

	// The time range selector for the time graphs in the Control Panel
	$(".timeRangeSelector").change(function(){
		console.log("Changed the time range to " + this.options[this.selectedIndex].getAttribute("val"));
		window.location = this.form.getAttribute("action")+"?timeRange=" + this.options[this.selectedIndex].getAttribute("val"); // Do a redirect as a form submission refuses to work properly
	});

	$(".unix_to_24_hour_time").each(function(){
		let unixTime = this.innerText;
		let date = new Date(unixTime*1000);
		console.log("date: ", date);
		let minutes = "0" + date.getMinutes();
		let formattedTime = date.getHours() + ":" + minutes.substr(-2);
		console.log("formattedTime:", formattedTime);
		this.innerText = formattedTime;
	});

	$(".unix_to_date").each(function(){
		// TODO: Localise this
		let monthList = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'];
		let date = new Date(this.innerText * 1000);
		console.log("date: ", date);
		let day = "0" + date.getDate();
		let formattedTime = monthList[date.getMonth()] + " " + day.substr(-2) + " " + date.getFullYear();
		console.log("formattedTime:", formattedTime);
		this.innerText = formattedTime;
	});

	this.onkeyup = function(event) {
		if(event.which == 37) this.querySelectorAll("#prevFloat a")[0].click();
		if(event.which == 39) this.querySelectorAll("#nextFloat a")[0].click();
	};

	function addPollInput() {
		console.log("clicked on pollinputinput");
		let dataPollInput = $(this).parent().attr("data-pollinput");
		console.log("dataPollInput: ", dataPollInput);
		if(dataPollInput == undefined) return;
		if(dataPollInput != (pollInputIndex-1)) return;

		$(".poll_content_row .formitem").append("<div class='pollinput' data-pollinput='"+pollInputIndex+"'><input type='checkbox' disabled /><label class='pollinputlabel'></label><input form='quick_post_form' name='pollinputitem["+pollInputIndex+"]' class='pollinputinput' type='text' placeholder='Add new poll option' /></div>");
		pollInputIndex++;
		console.log("new pollInputIndex: ", pollInputIndex);
		$(".pollinputinput").off("click");
		$(".pollinputinput").click(addPollInput);
	}

	var pollInputIndex = 1;
	$("#add_poll_button").click((event) => {
		event.preventDefault();
		$(".poll_content_row").removeClass("auto_hide");
		$("#has_poll_input").val("1");
		$(".pollinputinput").click(addPollInput);
	});

	//id="poll_results_{{.Poll.ID}}" class="poll_results auto_hide"
	$(".poll_results_button").click(function(){
		let pollID = $(this).attr("data-poll-id");
		$("#poll_results_" + pollID).removeClass("auto_hide");
		fetch("/poll/results/" + pollID, {
			credentials: 'same-origin'
		}).then((response) => response.text()).catch((error) => console.error("Error:",error)).then((rawData) => {
			// TODO: Make sure the received data is actually a list of integers
			let data = JSON.parse(rawData);

			let allZero = true;
			for(let i = 0; i < data.length; i++) {
				if(data[i] != "0") allZero = false;
			}
			if(allZero) {
				$("#poll_results_" + pollID + " .poll_no_results").removeClass("auto_hide");
				console.log("all zero")
				return;
			}

			$("#poll_results_" + pollID + " .user_content").html("<div id='poll_results_chart_"+pollID+"'></div>");
			console.log("rawData: ", rawData);
			console.log("series: ", data);
			Chartist.Pie('#poll_results_chart_' + pollID, {
 				series: data,
			}, {
				height: '120px',
			});
		})
	});

	runInitHook("end_init");
};
