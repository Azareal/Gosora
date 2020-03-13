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
var noAlerts = false;

// Topic move
var forumToMoveTo = 0;

function pushNotice(msg) {
	let aBox = document.getElementsByClassName("alertbox")[0];
	let div = document.createElement('div');
	div.innerHTML = Template_notice(msg).trim();
	aBox.appendChild(div);
	runInitHook("after_notice");
}

// TODO: Write a friendlier error handler which uses a .notice or something, we could have a specialised one for alerts
function ajaxError(xhr,status,errstr) {
	console.log("The AJAX request failed");
	console.log("xhr", xhr);
	console.log("status", status);
	console.log("err", errstr);
	if(status=="parsererror") console.log("The server didn't respond with a valid JSON response");
	console.trace();
}

function postLink(ev) {
	ev.preventDefault();
	let formAction = $(ev.target).closest('a').attr("href");
	$.ajax({ url: formAction, type: "POST", dataType: "json", error: ajaxError, data: {js: 1} });
}

function bindToAlerts() {
	console.log("bindToAlerts");
	$(".alertItem.withAvatar a").unbind("click");
	$(".alertItem.withAvatar a").click(function(ev) {
		ev.stopPropagation();
		ev.preventDefault();
		$.ajax({
			url: "/api/?a=set&m=dismiss-alert",
			type: "POST",
			dataType: "json",
			data: { id: $(this).attr("data-asid") },
			//error: ajaxError,
			success: () => {
				window.location.href = this.getAttribute("href");
			}
		});
	});
}

function addAlert(msg, notice=false) {
	var mmsg = msg.msg;
	if(mmsg[0]==".") mmsg = phraseBox["alerts"]["alerts"+mmsg];
	if("sub" in msg) {
		for(var i = 0; i < msg.sub.length; i++) mmsg = mmsg.replace("\{"+i+"\}", msg.sub[i]);
	}

	let aItem = Template_alert({
		ASID: msg.id,
		Path: msg.path,
		Avatar: msg.img || "",
		Message: mmsg
	})
	//alertMapping[msg.id] = aItem;
	let div = document.createElement('div');
	div.innerHTML = aItem.trim();
	alertMapping[msg.id] = div.firstChild;
	alertList.push(msg.id);

	if(notice) {
		// TODO: Add some sort of notification queue to avoid flooding the end-user with notices?
		// TODO: Use the site name instead of "Something Happened"
		if(Notification.permission === "granted") {
			var n = new Notification("Something Happened",{
				body: mmsg,
				icon: msg.img,
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
	console.log("alertCount",alertCount)
	runInitHook("after_update_alert_list", alertCount);
}

function setAlertError(menuAlerts,msg) {
	let alertListNode = menuAlerts.getElementsByClassName("alertList")[0];
	alertListNode.innerHTML = "<div class='alertItem'>"+msg+"</div>";
}

var alertsInitted = false;
var lastTc = 0;
function loadAlerts(menuAlerts, eTc=false) {
	if(!alertsInitted) return;
	let tc = "";
	if(eTc && lastTc != 0) tc = "&t=" + lastTc + "&c=" + alertCount;
	$.ajax({
		type:'get',
		dataType:'json',
		url:'/api/?m=alerts' + tc,
		success: data => {
			if("errmsg" in data) {
				setAlertError(menuAlerts,data.errmsg)
				return;
			}
			alertList = [];
			alertMapping = {};
			if(!data.hasOwnProperty("msgs")) data = {"msgs":[],"count":alertCount,"tc":lastTc};
			/*else if(data.count != (alertCount + data.msgs.length)) tc = false;
			if(eTc && lastTc != 0) {
				for(var i in data.msgs) wsAlertEvent(data.msgs[i]);
			} else {*/
				console.log("data",data);
				for(var i in data.msgs) addAlert(data.msgs[i]);
				alertCount = data.count;
				updateAlertList(menuAlerts);
			//}
			lastTc = data.tc;
		},
		error: (magic,theStatus,err) => {
			let errtxt = "Unable to get the alerts";
			try {
				var data = JSON.parse(magic.responseText);
				if("errmsg" in data) errtxt = data.errmsg;
			} catch(err) {
				console.log(magic.responseText);
				console.log(err);
			}
			console.log("err", err);
			setAlertError(menuAlerts,errtxt);
		}
	});
}

function SplitN(data,ch,n) {
	var out = [];
	if(data.length === 0) return out;

	var lastI = 0;
	var j = 0;
	var lastN = 1;
	for(let i = 0; i < data.length; i++) {
		if(data[i] === ch) {
			out[j++] = data.substring(lastI,i);
			lastI = i;
			if(lastN === n) break;
			lastN++;
		}
	}
	if(data.length > lastI) out[out.length - 1] += data.substring(lastI);
	return out;
}

function wsAlertEvent(data) {
	console.log("wsAlertEvent",data)
	addAlert(data, true);
	alertCount++;

	let aTmp = alertList;
	alertList = [alertList[alertList.length-1]];
	aTmp = aTmp.slice(0,-1);
	for(let i = 0; i < aTmp.length; i++) alertList.push(aTmp[i]);
	// TODO: Add support for other alert feeds like PM Alerts
	var generalAlerts = document.getElementById("general_alerts");
	// TODO: Make sure we update alertCount here
	lastTc = 0;
	updateAlertList(generalAlerts/*, alist*/);
}

function runWebSockets(resume=false) {
	let s = "";
	if(window.location.protocol == "https:") s = "s";
	conn = new WebSocket("ws"+s+"://" + document.location.host + "/ws/");

	conn.onerror = (err) => {
		console.log(err);
	}

	// TODO: Sync alerts, topic list, etc.
	conn.onopen = () => {
		console.log("The WebSockets connection was opened");
		if(resume) conn.send("resume " + document.location.pathname + " " + Math.round((new Date()).getTime() / 1000) + '\r');
		else conn.send("page " + document.location.pathname + '\r');
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
			if(!noAlerts) {
				let alertMenuList = document.getElementsByClassName("menu_alerts");
				for(var i=0; i < alertMenuList.length; i++) loadAlerts(alertMenuList[i]);
			}
			runWebSockets(true);
		}, backoff * 60 * 1000);

		if(wsBackoff > 0) {
			if(wsBackoff <= 5) setTimeout(() => wsBackoff--, 5.5 * 60 * 1000);
			else if(wsBackoff <= 12) setTimeout(() => wsBackoff--, 11.5 * 60 * 1000);
			else setTimeout(() => wsBackoff--, 20 * 60 * 1000);
		}
	}

	conn.onmessage = (event) => {
		if(!noAlerts && event.data[0] == "{") {
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
						if(key!=data.id) return;
						alertCount--;
						let index = -1;
						for(var i=0; i < alertList.length; i++) {
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
						let generalAlerts = document.getElementById("general_alerts");
						if(alertList.length < 8) loadAlerts(generalAlerts);
						else updateAlertList(generalAlerts);
					});
				}
			} else if("Topics" in data) {
				console.log("topic in data");
				console.log("data",data);
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
				for(let i=0; i < moreTopicBlocks.length; i++) {
					let moreTopicBlock = moreTopicBlocks[i];
					moreTopicBlock.classList.remove("more_topic_block_initial");
					moreTopicBlock.classList.add("more_topic_block_active");

					console.log("phraseBox",phraseBox);
					let msgBox = moreTopicBlock.getElementsByClassName("more_topics")[0];
					msgBox.innerText = phraseBox["topic_list"]["topic_list.changed_topics"].replace("%d",moreTopicCount);
				}
			} else {
				console.log("unknown message");
				console.log(data);
			}
		}

		let messages = event.data.split('\r');
		for(var i=0; i < messages.length; i++) {
			let message = messages[i];
			//console.log("message",message);
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

// TODO: Surely, there's a prettier and more elegant way of doing this?
function getExt(name) {
	if(!name.indexOf('.' > -1)) throw("This file doesn't have an extension");
	return name.split('.').pop();
}

(() => {
	addInitHook("pre_init", () => {
		runInitHook("pre_global");
		console.log("before notify on alert")
		// We can only get away with this because template_alert has no phrases, otherwise it too would have to be part of the "dance", I miss Go concurrency :(
		if(!noAlerts) {
		notifyOnScriptW("template_alert", e => {
			if(e!=undefined) console.log("failed alert? why?",e)
		}, () => {
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
		} else {
			addInitHook("after_phrases", () => {
				$(document).ready(() => {
					if(window["WebSocket"]) runWebSockets();
				});
			});
		}

		$(document).ready(mainInit);
	});
})();

// TODO: Use these in .filter_item and pass back an item count from the backend to work with here
// Ported from common/parser.go
function PageOffset(count, page, perPage) {
	let offset = 0;
	let lastPage = LastPage(count, perPage)
	if(page > 1) offset = (perPage * page) - perPage;
	else if (page == -1) {
		page = lastPage;
		offset = (perPage * page) - perPage;
	} else page = 1;

	// We don't want the offset to overflow the slices, if everything's in memory
	//if(offset >= (count - 1)) offset = 0;
	return {Offset:offset, Page:page, LastPage:lastPage};
}
function LastPage(count, perPage) {
	return (count / perPage) + 1
}
function Paginate(currentPage, lastPage, maxPages) {
	let diff = lastPage - currentPage;
	let pre = 3;
	if(diff < 3) pre = maxPages - diff;
	
	let page = currentPage - pre;
	if(page < 0) page = 0;
	let out = [];
	while(out.length < maxPages && page < lastPage){
		page++;
		out.push(page);
	}
	return out;
}

function mainInit(){
	runInitHook("start_init");

	$(".more_topics").click(ev => {
		ev.preventDefault();
		let moreTopicBlocks = document.getElementsByClassName("more_topic_block_active");
		for(let i = 0; i < moreTopicBlocks.length; i++) {
			let block = moreTopicBlocks[i];
			block.classList.remove("more_topic_block_active");
			block.classList.add("more_topic_block_initial");
		}
		$(".ajax_topic_dupe").fadeOut("slow", function(){
			$(this).remove();
		});
		$(".hide_ajax_topic").removeClass("hide_ajax_topic"); // TODO: Do Fade
		moreTopicCount = 0;
	})

	$(".add_like,.remove_like").click(function(ev) {
		ev.preventDefault();
		//$(this).unbind("click");
		let target = this.closest("a").getAttribute("href");
		console.log("target",target);

		let controls = this.closest(".controls");
		let hadLikes = controls.classList.contains("has_likes");
		let likeCountNode = controls.getElementsByClassName("like_count")[0];
		console.log("likeCountNode",likeCountNode);
		if(this.classList.contains("add_like")) {
			this.classList.remove("add_like");
			this.classList.add("remove_like");
			if(!hadLikes) controls.classList.add("has_likes");
			this.closest("a").setAttribute("href", target.replace("like","unlike"));
			likeCountNode.innerHTML = parseInt(likeCountNode.innerHTML) + 1;
		} else {
			this.classList.remove("remove_like");
			this.classList.add("add_like");
			let likeCount = parseInt(likeCountNode.innerHTML);
			if(likeCount==1) controls.classList.remove("has_likes");
			this.closest("a").setAttribute("href", target.replace("unlike","like"));
			likeCountNode.innerHTML = parseInt(likeCountNode.innerHTML) - 1;
		}

		//let likeButton = this;
		$.ajax({
			url: target,
			type: "POST",
			dataType: "json",
			data: { js: 1 },
			error: ajaxError,
			success: function (data,status,xhr) {
				if("success" in data && data["success"] == "1") return;
				// addNotice("Failed to add a like: {err}")
				//likeCountNode.innerHTML = parseInt(likeCountNode.innerHTML)-1;
				console.log("data",data);
				console.log("status",status);
				console.log("xhr",xhr);
			}
		});
	});

	$(".link_label").click(function(ev) {
		ev.preventDefault();
		let linkSelect = $('#'+$(this).attr("data-for"));
		if(!linkSelect.hasClass("link_opened")) {
			ev.stopPropagation();
			linkSelect.addClass("link_opened");
		}
	});

	function rebuildPaginator(lastPage) {
		let urlParams = new URLSearchParams(window.location.search);
		let page = urlParams.get('page');
		if(page=="") page = 1;

		let pageList = Paginate(page,lastPage,5)
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
				.then(resp => {
					if(!resp.ok) throw(url+q+"&js=1 failed to load");
					return resp.json();
				}).then(data => {
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
				}).catch(ex => {
					console.log("Unable to get script '"+url+q+"&js=1"+"'");
					console.log("ex", ex);
					console.trace();
				});
		});
	}

	// TODO: Render a headless topics.html instead of the individual topic rows and a bit of JS glue
	$(".filter_item").click(function(ev) {
		if(!window.location.pathname.startsWith("/topics/")) return
		ev.preventDefault();
		let that = this;
		let fid = this.getAttribute("data-fid");
		// TODO: Take mostviewed into account
		let url = "//"+window.location.host+"/topics/?fids="+fid;

		fetch(url+"&js=1", {credentials: "same-origin"})
		.then(resp => {
			if(!resp.ok) throw(url+"&js=1 failed to load");
			return resp.json();
		}).then(data => {
			console.log("data",data);
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
		}).catch(ex => {
			console.log("Unable to get script '"+url+"&js=1"+"'");
			console.log("ex",ex);
			console.trace();
		});
	});

	if (document.getElementById("topicsItemList")!==null) rebindPaginator();
	if (document.getElementById("forumItemList")!==null) rebindPaginator();

	// TODO: Show a search button when JS is disabled?
	$(".widget_search_input").keypress(function(e) {
		if(e.keyCode!='13') return;
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
			.then(resp => {
				if(!resp.ok) throw(url+q+"&js=1 failed to load");
				return resp.json();
			}).then(data => {
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
		}).catch(ex => {
			console.log("Unable to get script '"+url+q+"&js=1"+"'");
			console.log("ex",ex);
			console.trace();
		});
	});

	bindPage();

	$(".edit_field").click(function(ev) {
		ev.preventDefault();
		let bp = $(this).closest('.editable_parent');
		let block = bp.find('.editable_block').eq(0);
		block.html("<input name='edit_field' value='" + block.text() + "' type='text'/><a href='" + $(this).closest('a').attr("href") + "'><button class='submit_edit' type='submit'>Update</button></a>");

		$(".submit_edit").click(function(ev) {
			ev.preventDefault();
			let bp = $(this).closest('.editable_parent');
			let block = bp.find('.editable_block').eq(0);
			let content = block.find('input').eq(0).val();
			block.html(content);

			let formAction = $(this).closest('a').attr("href");
			$.ajax({
				url: formAction + "?s=" + me.User.S,
				type: "POST",
				dataType: "json",
				error: ajaxError,
				data: { js: 1, edit_item: content }
			});
		});
	});

	$(".edit_fields").click(function(ev) {
		ev.preventDefault();
		if($(this).find("input").length!==0) return;
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

		$(".submit_edit").click(function(ev) {
			ev.preventDefault();
			var outData = {js: 1}
			var bp = $(this).closest('.editable_parent');
			bp.find('.editable_block').each(function() {
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
			//console.log("Form Action", formAction);
			//console.log(outData);
			$.ajax({ url: formAction + "?s=" + me.User.S, type:"POST", dataType:"json", data: outData, error: ajaxError });
			bp.find('.hide_on_edit').removeClass("edit_opened");
			bp.find('.show_on_edit').removeClass("edit_opened");
		});
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
	$(".menu_alerts").click(function(ev) {
		ev.stopPropagation();
		if($(this).hasClass("selectedAlert")) return;
		if(!conn) loadAlerts(this);
		this.className += " selectedAlert";
		document.getElementById("back").className += " alertActive"
	});
	$(".link_select").click(ev => ev.stopPropagation());

	$("input,textarea,select,option").keyup(ev => ev.stopPropagation())

	$(".create_topic_link").click(ev => {
		ev.preventDefault();
		$(".topic_create_form").removeClass("auto_hide");
	});
	$(".topic_create_form .close_form").click(ev => {
		ev.preventDefault();
		$(".topic_create_form").addClass("auto_hide");
	});

	$("#themeSelectorSelect").change(function(){
		console.log("Changing the theme to " + this.options[this.selectedIndex].getAttribute("value"));
		$.ajax({
			url: this.form.getAttribute("action") + "?s=" + me.User.S,
			type: "POST",
			dataType: "json",
			data: { "theme": this.options[this.selectedIndex].getAttribute("value"), js: 1 },
			error: ajaxError,
			success: function (data,status,xhr) {
				console.log("Theme successfully switched");
				console.log("data",data);
				console.log("status",status);
				console.log("xhr",xhr);
				window.location.reload();
			}
		});
	});

	// The time range selector for the time graphs in the Control Panel
	$(".autoSubmitRedirect").change(function(){
		let elems = this.form.elements;
		let s = "";
		for(let i = 0; i < elems.length; i++) {
			let elem = elems[i];
			if(elem.nodeName=="SELECT") {
				s += elem.name + "=" + elem.options[elem.selectedIndex].getAttribute("value") + "&";
			}
			// TODO: Implement other element types...
		}
		if(s.length > 0) s = "?" + s.substr(0, s.length-1);

		window.location = this.form.getAttribute("action") + s; // Do a redirect as a form submission refuses to work properly
	});

	$(".unix_to_24_hour_time").each(function(){
		let unixTime = this.innerText;
		let date = new Date(unixTime*1000);
		console.log("date", date);
		let minutes = "0" + date.getMinutes();
		let formattedTime = date.getHours() + ":" + minutes.substr(-2);
		console.log("formattedTime", formattedTime);
		this.innerText = formattedTime;
	});

	$(".unix_to_date").each(function(){
		// TODO: Localise this
		let monthList = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'];
		let date = new Date(this.innerText * 1000);
		console.log("date", date);
		let day = "0" + date.getDate();
		let formattedTime = monthList[date.getMonth()] + " " + day.substr(-2) + " " + date.getFullYear();
		console.log("formattedTime",formattedTime);
		this.innerText = formattedTime;
	});

	$("spoiler").addClass("hide_spoil");
	$(".hide_spoil").click(function(ev) {
		ev.stopPropagation();
		ev.preventDefault();
		$(this).removeClass("hide_spoil");
		$(this).unbind("click");
	});

	this.onkeyup = function(ev) {
		if(ev.which == 37) this.querySelectorAll("#prevFloat a")[0].click();
		if(ev.which == 39) this.querySelectorAll("#nextFloat a")[0].click();
	};

	function asyncGetSheet(src) {
		return new Promise((resolve,reject) => {
			let res = document.createElement('link');
			res.async = true;
	
			const onloadHandler = (e,isAbort) => {
				if (isAbort || !res.readyState || /loaded|complete/.test(res.readyState)) {
					res.onload = null;
					res.onreadystatechange = null;
					res = undefined;
	
					isAbort ? reject(e) : resolve();
				}
			}
	
			res.onerror = (e) => {
				reject(e);
			};
			res.onload = onloadHandler;
			res.onreadystatechange = onloadHandler;
			res.href = src;
			res.rel = "stylesheet";
			res.type = "text/css";
	
			const prior = document.getElementsByTagName('link')[0];
			prior.parentNode.insertBefore(res,prior);
		});
	}

	function stripQ(name) {
		return name.split('?')[0];
	}

	$(".rowtopic a,a.rowtopic").click(function(ev) {
		let base = this.getAttribute("href");
		let href = base + "?i=1";
		fetch(href, {credentials:"same-origin"})
			.then(resp => {
				if(!resp.ok) throw(href+" failed to load");
				let xRes = resp.headers.get("x-res")
				for(let res of xRes.split(",")) {
					let pro;
					if(stripQ(getExt(res)) == "css") pro = asyncGetSheet("/s/"+res)
					else pro = asyncGetScript("/s/"+res)
						pro.then(() => console.log("Loaded " + res))
						.catch(e => {
							console.log("Unable to get res '"+res+"'");
							console.log("e",e);
							console.trace();
						});
				}
				return resp.text();
			}).then(data => {
				document.querySelector("#back").outerHTML = data;
				unbindTopic();
				bindTopic();
				$(".elapsed").remove();
				let obj = {Title: document.title, Url: base};
				history.pushState(obj, obj.Title, obj.Url);
			}).catch(ex => {
				console.log("Unable to get script '"+href+""+"'");
				console.log("ex",ex);
				console.trace();
			});
		
		ev.stopPropagation();
		ev.preventDefault();
	})

	runInitHook("almost_end_init");
	runInitHook("end_init");
}

function bindPage() {
	bindTopic();
}

function unbindPage() {
	unbindTopic();
}

function bindTopic() {
	$(".open_edit").click(ev => {
		ev.preventDefault();
		$('.hide_on_edit').addClass("edit_opened");
		$('.show_on_edit').addClass("edit_opened");
		runHook("open_edit");
	});
	
	$(".topic_item .submit_edit").click(function(ev){
		ev.preventDefault();
		let nameInput = $(".topic_name_input").val();
		$(".topic_name").html(nameInput);
		$(".topic_name").attr(nameInput);
		let contentInput = $('.topic_content_input').val();
		$(".topic_content").html(quickParse(contentInput));
		let statusInput = $('.topic_status_input').val();
		$(".topic_status_e:not(.open_edit)").html(statusInput);

		$('.hide_on_edit').removeClass("edit_opened");
		$('.show_on_edit').removeClass("edit_opened");
		runHook("close_edit");

		$.ajax({
			url: this.form.getAttribute("action"),
			type:"POST",
			dataType:"json",
			data: {
				name: nameInput,
				status: statusInput,
				content: contentInput,
				js: 1
			},
			error: ajaxError,
			success: (data,status,xhr) => {
				if("Content" in data) $(".topic_content").html(data["Content"]);
			}
		});
	});
	
	$(".delete_item").click(function(ev) {
		postLink(ev);
		$(this).closest('.deletable_block').remove();
	});

	// Miniature implementation of the parser to avoid sending as much data back and forth
	function quickParse(m) {
		m = m.replace(":)", "😀")
		m = m.replace(":(", "😞")
		m = m.replace(":D", "😃")
		m = m.replace(":P", "😛")
		m = m.replace(":O", "😲")
		m = m.replace(":p", "😛")
		m = m.replace(":o", "😲")
		m = m.replace(";)", "😉")
		m = m.replace("\n","<br>")
		return m
	}

	$(".edit_item").click(function(ev){
		ev.preventDefault();

		let bp = this.closest('.editable_parent');
		$(bp).find('.hide_on_edit').addClass("edit_opened");
		$(bp).find('.show_on_edit').addClass("edit_opened");
		$(bp).find('.hide_on_block_edit').addClass("edit_opened");
		$(bp).find('.show_on_block_edit').addClass("edit_opened");
		let srcNode = bp.querySelector(".edit_source");
		let block = bp.querySelector('.editable_block');
		block.classList.add("in_edit");

		let src = "";
		if(srcNode!=null) src = srcNode.innerText;
		else src = block.innerHTML;
		block.innerHTML = Template_topic_c_edit_post({
			ID: bp.getAttribute("id").slice("post-".length),
			Source: src,
			Ref: this.closest('a').getAttribute("href")
		})
		runHook("edit_item_pre_bind");

		$(".submit_edit").click(function(ev){
			ev.preventDefault();
			$(bp).find('.hide_on_edit').removeClass("edit_opened");
			$(bp).find('.show_on_edit').removeClass("edit_opened");
			$(bp).find('.hide_on_block_edit').removeClass("edit_opened");
			$(bp).find('.show_on_block_edit').removeClass("edit_opened");
			block.classList.remove("in_edit");
			let content = block.querySelector('textarea').value;
			block.innerHTML = quickParse(content);
			if(srcNode!=null) srcNode.innerText = content;

			let formAction = this.closest('a').getAttribute("href");
			// TODO: Bounce the parsed post back and set innerHTML to it?
			$.ajax({
				url: formAction,
				type:"POST",
				dataType:"json",
				data: { js: 1, edit_item: content },
				error: ajaxError,
				success: (data,status,xhr) => {
					if("Content" in data) block.innerHTML = data["Content"];
				}
			});
		});
	});

	$(".quote_item").click(function(ev){
		ev.preventDefault();
		ev.stopPropagation();
		let src = this.closest(".post_item").getElementsByClassName("edit_source")[0];
		let content = document.getElementById("input_content")
		console.log("content.value", content.value);

		let item;
		if(content.value=="") item = "<blockquote>" + src.innerHTML + "</blockquote>"
		else item = "\r\n<blockquote>" + src.innerHTML + "</blockquote>";
		content.value = content.value + item;
		console.log("content.value", content.value);

		// For custom / third party text editors
		quoteItemCallback(src.innerHTML,item);
	});

	//id="poll_results_{{.Poll.ID}}" class="poll_results auto_hide"
	$(".poll_results_button").click(function(){
		let pollID = $(this).attr("data-poll-id");
		$("#poll_results_" + pollID).removeClass("auto_hide");
		fetch("/poll/results/" + pollID, {
			credentials: 'same-origin'
		}).then(resp => resp.text()).catch(err => console.error("err",err)).then(rawData => {
			// TODO: Make sure the received data is actually a list of integers
			let data = JSON.parse(rawData);
			let allZero = true;
			for(let i = 0; i < data.length; i++) {
				if(data[i] != "0") allZero = false;
			}
			if(allZero) {
				$("#poll_results_"+pollID+" .poll_no_results").removeClass("auto_hide");
				console.log("all zero")
				return;
			}

			$("#poll_results_"+pollID+" .user_content").html("<div id='poll_results_chart_"+pollID+"'></div>");
			console.log("rawData",rawData);
			console.log("series",data);
			Chartist.Pie('#poll_results_chart_'+pollID, {
 				series: data,
			}, {
				height: '120px',
			});
		})
	});

	runHook("end_bind_topic");
}

function unbindTopic() {
	$(".open_edit").unbind("click");
	$(".topic_item .submit_edit").unbind("click");
	$(".delete_item").unbind("click");
	$(".edit_item").unbind("click");
	$(".submit_edit").unbind("click");
	$(".quote_item").unbind("click");
	$(".poll_results_button").unbind("click");
	runHook("end_unbind_topic");
}