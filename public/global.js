'use strict';
var form_vars = {};
var alertList = [];
var alertCount = 0;
var conn;
var selectedTopics = [];
var attachItemCallback = function(){}

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

// TODO: Add the ability for users to dismiss alerts
function loadAlerts(menuAlerts)
{
	var alertListNode = menuAlerts.getElementsByClassName("alertList")[0];
	var alertCounterNode = menuAlerts.getElementsByClassName("alert_counter")[0];
	alertCounterNode.textContent = "0";
	$.ajax({
		type: 'get',
		dataType: 'json',
		url:'/api/?action=get&module=alerts',
		success: function(data) {
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

				if("avatar" in msg) {
					alist += "<div class='alertItem withAvatar' style='background-image:url(\""+msg.avatar+"\");'><a class='text' data-asid='"+msg.asid+"' href=\""+msg.path+"\">"+mmsg+"</a></div>";
					alertList.push("<div class='alertItem withAvatar' style='background-image:url(\""+msg.avatar+"\");'><a class='text' data-asid='"+msg.asid+"' href=\""+msg.path+"\">"+mmsg+"</a></div>");
				} else {
					alist += "<div class='alertItem'><a href=\""+msg.path+"\" class='text'>"+mmsg+"</a></div>";
					alertList.push("<div class='alertItem'><a href=\""+msg.path+"\" class='text'>"+mmsg+"</a></div>");
				}
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
		error: function(magic,theStatus,error) {
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

function runWebSockets() {
	if(window.location.protocol == "https:")
		conn = new WebSocket("wss://" + document.location.host + "/ws/");
	else conn = new WebSocket("ws://" + document.location.host + "/ws/");

	conn.onopen = function() {
		console.log("The WebSockets connection was opened");
		conn.send("page " + document.location.pathname + '\r');
		// TODO: Don't ask again, if it's denied. We could have a setting in the UCP which automatically requests this when someone flips desktop notifications on
		Notification.requestPermission();
	}
	conn.onclose = function() {
		conn = false;
		console.log("The WebSockets connection was closed");
	}
	conn.onmessage = function(event) {
		//console.log("WS_Message:", event.data);
		if(event.data[0] == "{") {
			try {
				var data = JSON.parse(event.data);
			} catch(err) {
				console.log(err);
			}

			if ("msg" in data) {
				var msg = data.msg
				if("sub" in data)
					for(var i = 0; i < data.sub.length; i++)
						msg = msg.replace("\{"+i+"\}", data.sub[i]);

				if("avatar" in data) alertList.push("<div class='alertItem withAvatar' style='background-image:url(\""+data.avatar+"\");'><a class='text' data-asid='"+data.asid+"' href=\""+data.path+"\">"+msg+"</a></div>");
				else alertList.push("<div class='alertItem'><a href=\""+data.path+"\" class='text'>"+msg+"</a></div>");
				if(alertList.length > 8) alertList.shift();
				//console.log("post alertList",alertList);
				alertCount++;

				var alist = ""
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
		}

		var messages = event.data.split('\r');
		for(var i = 0; i < messages.length; i++) {
			//console.log("Message: ",messages[i]);
			if(messages[i].startsWith("set ")) {
				//msgblocks = messages[i].split(' ',3);
				let msgblocks = SplitN(messages[i]," ",3);
				if(msgblocks.length < 3) continue;
				document.querySelector(msgblocks[1]).innerHTML = msgblocks[2];
			} else if(messages[i].startsWith("set-class ")) {
				let msgblocks = SplitN(messages[i]," ",3);
				if(msgblocks.length < 3) continue;
				document.querySelector(msgblocks[1]).className = msgblocks[2];
			}
		}
	}
}

$(document).ready(function(){
	if(window["WebSocket"]) runWebSockets();
	else conn = false;

	$(".open_edit").click(function(event){
		//console.log("clicked on .open_edit");
		event.preventDefault();
		$(".hide_on_edit").hide();
		$(".show_on_edit").show();
	});

	$(".topic_item .submit_edit").click(function(event){
		event.preventDefault();
		//console.log("clicked on .topic_item .submit_edit");
		$(".topic_name").html($(".topic_name_input").val());
		$(".topic_content").html($(".topic_content_input").val());
		$(".topic_status_e:not(.open_edit)").html($(".topic_status_input").val());

		$(".hide_on_edit").show();
		$(".show_on_edit").hide();

		let topicNameInput = $('.topic_name_input').val();
		let topicStatusInput = $('.topic_status_input').val();
		let topicContentInput = $('.topic_content_input').val();
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

	$(".delete_item").click(function(event)
	{
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

	$("#forum_quick_perms").click(function(){
		$(".submit_edit").click(function(event){

		});
	});

	$(".edit_field").click(function(event)
	{
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
				url: formAction + "?session=" + session,
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
			if(fieldType=="list")
			{
				var fieldValue = this.getAttribute("data-value");
				if(fieldName in form_vars) var it = form_vars[fieldName];
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
			$.ajax({ url: formAction + "?session=" + session, type:"POST", dataType:"json", data: outData, error: ajaxError });
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

	$(this).click(function() {
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

	var alertMenuList = document.getElementsByClassName("menu_alerts");
	for(var i = 0; i < alertMenuList.length; i++) {
		loadAlerts(alertMenuList[i]);
	}

	$(".menu_alerts").click(function(event) {
		event.stopPropagation();
		if($(this).hasClass("selectedAlert")) return;
		if(!conn) loadAlerts(this);
		this.className += " selectedAlert";
		document.getElementById("back").className += " alertActive"
	});

	$("input,textarea,select,option").keyup(function(event){
		event.stopPropagation();
	})

	$(".create_topic_link").click(function(event){
		event.preventDefault();
		$(".topic_create_form").show();
	});
	$(".topic_create_form .close_form").click(function(){
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
		for(let i = 0; i < files.length; i++) {
			console.log("files[" + i + "]",files[i]);
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
					crypto.subtle.digest('SHA-256',e.target.result).then(function(hash) {
						const hashArray = Array.from(new Uint8Array(hash))
						return hashArray.map(b => ('00' + b.toString(16)).slice(-2)).join('')
					}).then(function(hash) {
						console.log("hash",hash);
						let content = document.getElementById("input_content")
						console.log("content.value", content.value);
						
						let attachItem;
						if(content.value == "") attachItem = "//" + siteURL + "/attachs/" + hash + "." + ext;
						else attachItem = "\r\n//" + siteURL + "/attachs/" + hash + "." + ext;
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
	}

	var uploadFiles = document.getElementById("upload_files");
	if(uploadFiles != null) {
		uploadFiles.addEventListener("change", uploadFileHandler, false);
	}
	
	$(".moderate_link").click(function(event) {
		event.preventDefault();
		$(".pre_opt").removeClass("auto_hide");
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
		$(".mod_floater_submit").click(function(event){
			event.preventDefault();
			let selectNode = this.form.querySelector(".mod_floater_options");
			let optionNode = selectNode.options[selectNode.selectedIndex];
			let action = optionNode.getAttribute("val");
			//console.log("action",action);

			// Handle these specially
			switch(action) {
				case "move":
					console.log("move action");
					$("#mod_topic_mover").removeClass("auto_hide");
					return;
			}
			
			let url = "/topic/"+action+"/submit/";
			//console.log("JSON.stringify(selectedTopics) ", JSON.stringify(selectedTopics));
			$.ajax({
				url: url,
				type: "POST",
				data: JSON.stringify(selectedTopics),
				contentType: "application/json",
				error: ajaxError,
				success: function() {
					window.location.reload();
				}
			});
		});
	});

	$("#themeSelectorSelect").change(function(){
		console.log("Changing the theme to " + this.options[this.selectedIndex].getAttribute("val"));
		$.ajax({
			url: this.form.getAttribute("action") + "?session=" + session,
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


});
