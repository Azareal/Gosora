'use strict';
var formVars = {};
var alertList = [];
var alertCount = 0;
var moreTopicCount = 0;
var conn;
var selectedTopics = [];
var attachItemCallback = function(){}

// Topic move
var forumToMoveTo = 0;

// TODO: Write a friendlier error handler which uses a .notice or something, we could have a specialised one for alerts
function ajaxError(xhr,status,errstr) {
	console.log("The AJAX request failed");
	console.log("xhr", xhr);
	console.log("status", status);
	console.log("errstr", errstr);
	if(status=="parsererror") {
		console.log("The server didn't respond with a valid JSON response");
	}
	console.trace();
}

function postLink(event)
{
	event.preventDefault();
	let formAction = $(event.target).closest('a').attr("href");
	//console.log("Form Action: " + formAction);
	$.ajax({ url: formAction, type: "POST", dataType: "json", error: ajaxError, data: {js: "1"} });
}

function bindToAlerts() {
	$(".alertItem.withAvatar a").click(function(event) {
		event.stopPropagation();
		$.ajax({ url: "/api/?action=set&module=dismiss-alert", type: "POST", dataType: "json", error: ajaxError, data: { asid: $(this).attr("data-asid") } });
	});
}

var alertsInitted = false;
// TODO: Add the ability for users to dismiss alerts
function loadAlerts(menuAlerts)
{
	if(!alertsInitted) return;

	var alertListNode = menuAlerts.getElementsByClassName("alertList")[0];
	var alertCounterNode = menuAlerts.getElementsByClassName("alert_counter")[0];
	alertCounterNode.textContent = "0";
	$.ajax({
		type: 'get',
		dataType: 'json',
		url:'/api/?action=get&module=alerts',
		success: (data) => {
			if("errmsg" in data) {
				alertListNode.innerHTML = "<div class='alertItem'>"+data.errmsg+"</div>";
				return;
			}

			var alist = "";
			for(var i in data.msgs) {
				var msg = data.msgs[i];
				var mmsg = msg.msg;
				if("sub" in msg) {
					for(var i = 0; i < msg.sub.length; i++) {
						mmsg = mmsg.replace("\{"+i+"\}", msg.sub[i]);
						//console.log("Sub #" + i + ":",msg.sub[i]);
					}
				}

				let aItem = Template_alert({
					ASID: msg.asid || 0,
					Path: msg.path,
					Avatar: msg.avatar || "",
					Message: mmsg
				})
				alist += aItem;
				alertList.push(aItem);
				//console.log(msg);
				//console.log(mmsg);
			}

			if(alist == "") alist = "<div class='alertItem'>You don't have any alerts</div>";
			alertListNode.innerHTML = alist;

			if(data.msgCount != 0 && data.msgCount != undefined) {
				alertCounterNode.textContent = data.msgCount;
				menuAlerts.classList.add("has_alerts");
			} else {
				menuAlerts.classList.remove("has_alerts");
			}
			alertCount = data.msgCount;

			bindToAlerts();
		},
		error: (magic,theStatus,error) => {
			let errtxt
			try {
				var data = JSON.parse(magic.responseText);
				if("errmsg" in data) errtxt = data.errmsg;
				else errtxt = "Unable to get the alerts";
			} catch(err) {
				errtxt = "Unable to get the alerts";
				console.log(magic.responseText);
				console.log(err);
			}
			console.log("error", error);
			alertListNode.innerHTML = "<div class='alertItem'>"+errtxt+"</div>";
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
	var msg = data.msg;
	if("sub" in data) {
		for(var i = 0; i < data.sub.length; i++) {
			msg = msg.replace("\{"+i+"\}", data.sub[i]);
		}
	}

	let aItem = Template_alert({
		ASID: data.asid || 0,
		Path: data.path,
		Avatar: data.avatar || "",
		Message: msg
	})
	alertList.push(aItem);
	if(alertList.length > 8) alertList.shift();
	//console.log("post alertList",alertList);
	alertCount++;

	var alist = "";
	for (var i = 0; i < alertList.length; i++) alist += alertList[i];

	//console.log(alist);
	// TODO: Add support for other alert feeds like PM Alerts
	var generalAlerts = document.getElementById("general_alerts");
	var alertListNode = generalAlerts.getElementsByClassName("alertList")[0];
	var alertCounterNode = generalAlerts.getElementsByClassName("alert_counter")[0];
	alertListNode.innerHTML = alist;
	alertCounterNode.textContent = alertCount;

	// TODO: Add some sort of notification queue to avoid flooding the end-user with notices?
	// TODO: Use the site name instead of "Something Happened"
	if(Notification.permission === "granted") {
		var n = new Notification("Something Happened",{
			body: msg,
			icon: data.avatar,
		});
		setTimeout(n.close.bind(n), 8000);
	}

	bindToAlerts();
}

function runWebSockets() {
	if(window.location.protocol == "https:") {
		conn = new WebSocket("wss://" + document.location.host + "/ws/");
	} else conn = new WebSocket("ws://" + document.location.host + "/ws/");

	conn.onerror = (err) => {
		console.log(err);
	}

	conn.onopen = () => {
		console.log("The WebSockets connection was opened");
		conn.send("page " + document.location.pathname + '\r');
		// TODO: Don't ask again, if it's denied. We could have a setting in the UCP which automatically requests this when someone flips desktop notifications on
		if(me.User.ID > 0) {
			Notification.requestPermission();
		}
	}

	conn.onclose = () => {
		conn = false;
		console.log("The WebSockets connection was closed");
	}

	conn.onmessage = (event) => {
		//console.log("WSMessage:", event.data);
		if(event.data[0] == "{") {
			console.log("json message");
			let data = "";
			try {
				data = JSON.parse(event.data);
			} catch(err) {
				console.log(err);
				return;
			}

			if ("msg" in data) {
				// TODO: Fix the data race where the alert template hasn't been loaded yet
				wsAlertEvent(data);
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
				document.querySelector(msgblocks[1]).innerHTML = msgblocks[2];
			} else if(message.startsWith("set-class ")) {
				document.querySelector(msgblocks[1]).className = msgblocks[2];
			}
		}
	}
}

(() => {
	addInitHook("pre_init", () => {
		// We can only get away with this because template_alert has no phrases, otherwise it too would have to be part of the "dance", I miss Go concurrency :(
		loadScript("template_alert.js", () => {
			console.log("Loaded template_alert.js");
			$(document).ready(() => {
				alertsInitted = true;
				var alertMenuList = document.getElementsByClassName("menu_alerts");
				for(var i = 0; i < alertMenuList.length; i++) {
					loadAlerts(alertMenuList[i]);
				}
			});
		});

		if(window["WebSocket"]) runWebSockets();
		else conn = false;

		$(document).ready(mainInit);
	});
})();

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
		let likeButton = this;
		let target = this.closest("a").getAttribute("href");
		console.log("target: ", target);
		likeButton.classList.remove("add_like");
		likeButton.classList.add("remove_like");
		let controls = likeButton.closest(".controls");
		let hadLikes = controls.classList.contains("has_likes");
		if(!hadLikes) controls.classList.add("has_likes");
		let likeCountNode = controls.getElementsByClassName("like_count")[0];
		console.log("likeCountNode",likeCountNode);
		likeCountNode.innerHTML = parseInt(likeCountNode.innerHTML) + 1;
		
		$.ajax({
			url: target,
			type: "POST",
			dataType: "json",
			data: { isJs: 1 },
			error: ajaxError,
			success: function (data, status, xhr) {
				if("success" in data) {
					if(data["success"] == "1") {
						return;
					}
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

	$(".open_edit").click((event) => {
		event.preventDefault();
		$(".hide_on_edit").hide();
		$(".show_on_edit").show();
	});

	$(".topic_item .submit_edit").click(function(event){
		event.preventDefault();
		let topicNameInput = $(".topic_name_input").val();
		$(".topic_name").html(topicNameInput);
		$(".topic_name").attr(topicNameInput);
		let topicContentInput = $('.topic_content_input').val();
		$(".topic_content").html(topicContentInput.replace(/(\n)+/g,"<br />"));
		let topicStatusInput = $('.topic_status_input').val();
		$(".topic_status_e:not(.open_edit)").html(topicStatusInput);

		$(".hide_on_edit").show();
		$(".show_on_edit").hide();

		let formAction = this.form.getAttribute("action");
		//console.log("New Topic Name: ", topicNameInput);
		//console.log("New Topic Status: ", topicStatusInput);
		//console.log("New Topic Content: ", topicContentInput);
		//console.log("Form Action: ", formAction);
		$.ajax({
			url: formAction,
			type: "POST",
			dataType: "json",
			error: ajaxError,
			data: {
				topic_name: topicNameInput,
				topic_status: topicStatusInput,
				topic_content: topicContentInput,
				topic_js: 1
			}
		});
	});

	$(".delete_item").click(function(event) {
		postLink(event);
		$(this).closest('.deletable_block').remove();
	});

	$(".edit_item").click(function(event){
		event.preventDefault();
		let blockParent = $(this).closest('.editable_parent');
		let block = blockParent.find('.editable_block').eq(0);
		block.html("<textarea style='width: 99%;' name='edit_item'>" + block.html() + "</textarea><br /><a href='" + $(this).closest('a').attr("href") + "'><button class='submit_edit' type='submit'>Update</button></a>");

		$(".submit_edit").click(function(event){
			event.preventDefault();
			let blockParent = $(this).closest('.editable_parent');
			let block = blockParent.find('.editable_block').eq(0);
			let newContent = block.find('textarea').eq(0).val();
			block.html(newContent);

			var formAction = $(this).closest('a').attr("href");
			//console.log("Form Action:",formAction);
			$.ajax({ url: formAction, type: "POST", error: ajaxError, dataType: "json", data: { isJs: "1", edit_item: newContent }
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
			//console.log("Form Action:", formAction);
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
		//console.log(blockParent);
		blockParent.find('.hide_on_edit').hide();
		blockParent.find('.show_on_edit').show();
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
				//console.log("Field Name:",fieldName);
				//console.log("Field Type:",fieldType);
				//console.log("Field Value:",fieldValue);
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

		$(".submit_edit").click(function(event)
		{
			event.preventDefault();
			//console.log("running .submit_edit event");
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
			blockParent.find('.hide_on_edit').show();
			blockParent.find('.show_on_edit').hide();
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

	$(this).click(() => {
		$(".selectedAlert").removeClass("selectedAlert");
		$("#back").removeClass("alertActive");
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

	$("input,textarea,select,option").keyup(event => event.stopPropagation())

	$(".create_topic_link").click((event) => {
		event.preventDefault();
		$(".topic_create_form").show();
	});
	$(".topic_create_form .close_form").click((event) => {
		event.preventDefault();
		$(".topic_create_form").hide();
	});

	function uploadFileHandler() {
		var fileList = this.files;
		// Truncate the number of files to 5
		let files = [];
		for(var i = 0; i < fileList.length && i < 5; i++)
			files[i] = fileList[i];

		// Iterate over the files
		let totalSize = 0;
		for(let i = 0; i < files.length; i++) {
			console.log("files[" + i + "]",files[i]);
			totalSize += files[i]["size"];

			let reader = new FileReader();
			reader.onload = function(e) {
				var fileDock = document.getElementById("upload_file_dock");
				var fileItem = document.createElement("label");
				console.log("fileItem",fileItem);

				if(!files[i]["name"].indexOf('.' > -1)) {
					// TODO: Surely, there's a prettier and more elegant way of doing this?
					alert("This file doesn't have an extension");
					return;
				}

				var ext = files[i]["name"].split('.').pop();
				fileItem.innerText = "." + ext;
				fileItem.className = "formbutton uploadItem";
				fileItem.style.backgroundImage = "url("+e.target.result+")";

				fileDock.appendChild(fileItem);

				let reader = new FileReader();
				reader.onload = function(e) {
					crypto.subtle.digest('SHA-256',e.target.result)
						.then(function(hash) {
							const hashArray = Array.from(new Uint8Array(hash))
							return hashArray.map(b => ('00' + b.toString(16)).slice(-2)).join('')
						}).then(function(hash) {
							console.log("hash",hash);
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
				}
				reader.readAsArrayBuffer(files[i]);
			}
			reader.readAsDataURL(files[i]);
		}
		if(totalSize > me.Site.MaxRequestSize) {
			// TODO: Use a notice instead
			alert("You can't upload this much data at once, max: " + me.Site.MaxRequestSize);
		}
	}

	var uploadFiles = document.getElementById("upload_files");
	if(uploadFiles != null) {
		uploadFiles.addEventListener("change", uploadFileHandler, false);
	}
	
	$(".moderate_link").click((event) => {
		event.preventDefault();
		$(".pre_opt").removeClass("auto_hide");
		$(".moderate_link").addClass("moderate_open");
		$(".topic_row").each(function(){
			$(this).click(function(){
				selectedTopics.push(parseInt($(this).attr("data-tid"),10));
				if(selectedTopics.length==1) {
					$(".mod_floater_head span").html("What do you want to do with this topic?");
				} else {
					$(".mod_floater_head span").html("What do you want to do with these "+selectedTopics.length+" topics?");
				}
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
			//console.log("action", action);

			// Handle these specially
			switch(action) {
				case "move":
					console.log("move action");
					let modTopicMover = $("#mod_topic_mover");
					$("#mod_topic_mover").removeClass("auto_hide");
					$("#mod_topic_mover .pane_row").click(function(){
						modTopicMover.find(".pane_row").removeClass("pane_selected");
						let fid = this.getAttribute("data-fid");
						if (fid == null) {
							return;
						}
						this.classList.add("pane_selected");
						console.log("fid: " + fid);
						forumToMoveTo = fid;

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
		$("#poll_results_" + pollID + " .user_content").html("<div id='poll_results_chart_"+pollID+"'></div>");
		$("#poll_results_" + pollID).removeClass("auto_hide");
		fetch("/poll/results/" + pollID, {
			credentials: 'same-origin'
		}).then((response) => response.text()).catch((error) => console.error("Error:",error)).then((rawData) => {
			// TODO: Make sure the received data is actually a list of integers
			let data = JSON.parse(rawData);
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
