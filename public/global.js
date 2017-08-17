var form_vars = {};
var alertList = [];
var alertCount = 0;
var conn;

function post_link(event)
{
	event.preventDefault();
	var form_action = $(event.target).closest('a').attr("href");
	//console.log("Form Action: " + form_action);
	$.ajax({ url: form_action, type: "POST", dataType: "json", data: {js: "1"} });
}

function load_alerts(menu_alerts)
{
	var alertListNode = menu_alerts.getElementsByClassName("alertList")[0];
	var alertCounterNode = menu_alerts.getElementsByClassName("alert_counter")[0];
	alertCounterNode.textContent = "0";
	$.ajax({
			type: 'get',
			dataType: 'json',
			url:'/api/?action=get&module=alerts&format=json',
			success: function(data) {
				if("errmsg" in data) {
					alertListNode.innerHTML = "<div class='alertItem'>"+data.errmsg+"</div>";
					return;
				}

				var alist = "";
				var anyAvatar = false
				for(var i in data.msgs) {
					var msg = data.msgs[i];
					var mmsg = msg.msg;

					if("sub" in msg) {
						for(var i = 0; i < msg.sub.length; i++) {
							mmsg = mmsg.replace("\{"+i+"\}", msg.sub[i]);
							//console.log("Sub #" + i);
							//console.log(msg.sub[i]);
						}
					}

					if("avatar" in msg) {
						alist += "<div class='alertItem withAvatar' style='background-image:url(\""+msg.avatar+"\");'><a class='text' href=\""+msg.path+"\">"+mmsg+"</a></div>";
						alertList.push("<div class='alertItem withAvatar' style='background-image:url(\""+msg.avatar+"\");'><a class='text' href=\""+msg.path+"\">"+mmsg+"</a></div>");
						anyAvatar = true
					} else {
						alist += "<div class='alertItem'><a href=\""+msg.path+"\" class='text'>"+mmsg+"</a></div>";
						alertList.push("<div class='alertItem'><a href=\""+msg.path+"\" class='text'>"+mmsg+"</a></div>");
					}
					//console.log(msg);
					//console.log(mmsg);
				}

				if(alist == "") alist = "<div class='alertItem'>You don't have any alerts</div>";
				else {
					//menu_alerts.removeClass("hasAvatars");
					//if(anyAvatar) menu_alerts.addClass("hasAvatars");
				}
				alertListNode.innerHTML = alist;
				if(data.msgCount != 0 && data.msgCount != undefined) {
					alertCounterNode.textContent = data.msgCount;
					menu_alerts.classList.add("has_alerts");
				} else {
					menu_alerts.classList.remove("has_alerts");
				}
				alertCount = data.msgCount;
			},
			error: function(magic,theStatus,error) {
				try {
					var data = JSON.parse(magic.responseText);
					if("errmsg" in data) errtxt = data.errmsg;
					else errtxt = "Unable to get the alerts";
				} catch(err) {
					errtxt = "Unable to get the alerts";
					console.log(magic.responseText);
					console.log(err);
				}
				alertListNode.innerHTML = "<div class='alertItem'>"+errtxt+"</div>";
			}
		});
}

function SplitN(data,ch,n) {
	var out = [];
	if(data.length == 0) return out;

	var lastIndex = 0;
	var j = 0;
	var lastN = 1;
	for(var i = 0; i < data.length; i++) {
		if(data[i] == ch) {
			out[j++] = data.substring(lastIndex,i);
			lastIndex = i;
			if(lastN == n) break;
			lastN++;
		}
	}
	if(data.length > lastIndex) out[out.length - 1] += data.substring(lastIndex);
	return out;
}

$(document).ready(function(){
	if(window["WebSocket"]) {
		if(window.location.protocol == "https:")
			conn = new WebSocket("wss://" + document.location.host + "/ws/");
		else conn = new WebSocket("ws://" + document.location.host + "/ws/");

		conn.onopen = function() {
			conn.send("page " + document.location.pathname + '\r');
			// TO-DO: Don't ask again, if it's denied. We could have a setting in the UCP which automatically requests this when someone flips desktop notifications on
			Notification.requestPermission();
		}
		conn.onclose = function() {
			conn = false;
		}
		conn.onmessage = function(event) {
			//console.log("WS_Message: ",event.data);
			if(event.data[0] == "{") {
				try {
					var data = JSON.parse(event.data);
				} catch(err) { console.log(err); }

				if ("msg" in data) {
					var msg = data.msg
					if("sub" in data) {
						for(var i = 0; i < data.sub.length; i++) {
							msg = msg.replace("\{"+i+"\}", data.sub[i]);
						}
					}

					if("avatar" in data) alertList.push("<div class='alertItem withAvatar' style='background-image:url(\""+data.avatar+"\");'><a class='text' href=\""+data.path+"\">"+msg+"</a></div>");
					else alertList.push("<div class='alertItem'><a href=\""+data.path+"\" class='text'>"+msg+"</a></div>");
					if(alertList.length > 8) alertList.shift();
					//console.log("post alertList",alertList);
					alertCount++;

					var alist = ""
					for (var i = 0; i < alertList.length; i++) {
						alist += alertList[i];
					}

					//console.log(alist);
					$("#general_alerts").find(".alertList").html(alist); // Add support for other alert feeds like PM Alerts
					$("#general_alerts").find(".alert_counter").text(alertCount);

					// TO-DO: Add some sort of notification queue to avoid flooding the end-user with notices?
					// TO-DO: Use the site name instead of "Something Happened"
					if(Notification.permission === "granted") {
						var n = new Notification("Something Happened",{
							body: msg,
							icon: data.avatar,
						});
						setTimeout(n.close.bind(n), 8000);
					}
				}
			}

			var messages = event.data.split('\r');
			for(var i = 0; i < messages.length; i++) {
				//console.log("Message:");
				//console.log(messages[i]);
				if(messages[i].startsWith("set ")) {
					//msgblocks = messages[i].split(' ',3);
					msgblocks = SplitN(messages[i]," ",3);
					if(msgblocks.length < 3) continue;
					document.querySelector(msgblocks[1]).innerHTML = msgblocks[2];
				} else if(messages[i].startsWith("set-class ")) {
					msgblocks = SplitN(messages[i]," ",3);
					if(msgblocks.length < 3) continue;
					document.querySelector(msgblocks[1]).className = msgblocks[2];
				}
			}
		}
	}
	else conn = false;

	$(".open_edit").click(function(event){
		//console.log("Clicked on edit");
		event.preventDefault();
		$(".hide_on_edit").hide();
		$(".show_on_edit").show();
	});

	$(".topic_item .submit_edit").click(function(event){
		event.preventDefault();
		$(".topic_name").html($(".topic_name_input").val());
		$(".topic_content").html($(".topic_content_input").val());
		$(".topic_status_e:not(.open_edit)").html($(".topic_status_input").val());

		$(".hide_on_edit").show();
		$(".show_on_edit").hide();

		var topic_name_input = $('.topic_name_input').val();
		var topic_status_input = $('.topic_status_input').val();
		var topic_content_input = $('.topic_content_input').val();
		var form_action = $(this).closest('form').attr("action");
		//console.log("New Topic Name: " + topic_name_input);
		//console.log("New Topic Status: " + topic_status_input);
		//console.log("Form Action: " + form_action);
		$.ajax({
			url: form_action,
			type: "POST",
			dataType: "json",
			data: {
				topic_name: topic_name_input,
				topic_status: topic_status_input,
				topic_content: topic_content_input,
				topic_js: 1
			}
		});
	});

	$(".delete_item").click(function(event)
	{
		post_link(event);
		var block = $(this).closest('.deletable_block');
		block.remove();
	});

	$(".edit_item").click(function(event)
	{
		event.preventDefault();
		var block_parent = $(this).closest('.editable_parent');
		var block = block_parent.find('.editable_block').eq(0);
		block.html("<textarea style='width: 99%;' name='edit_item'>" + block.html() + "</textarea><br /><a href='" + $(this).closest('a').attr("href") + "'><button class='submit_edit' type='submit'>Update</button></a>");

		$(".submit_edit").click(function(event)
		{
			event.preventDefault();
			var block_parent = $(this).closest('.editable_parent');
			var block = block_parent.find('.editable_block').eq(0);
			var newContent = block.find('textarea').eq(0).val();
			block.html(newContent);

			var form_action = $(this).closest('a').attr("href");
			//console.log("Form Action: " + form_action);
			$.ajax({ url: form_action, type: "POST", dataType: "json", data: { is_js: "1", edit_item: newContent }
			});
		});
	});

	$(".edit_field").click(function(event)
	{
		event.preventDefault();
		var block_parent = $(this).closest('.editable_parent');
		var block = block_parent.find('.editable_block').eq(0);
		block.html("<input name='edit_field' value='" + block.text() + "' type='text'/><a href='" + $(this).closest('a').attr("href") + "'><button class='submit_edit' type='submit'>Update</button></a>");

		$(".submit_edit").click(function(event)
		{
			event.preventDefault();
			var block_parent = $(this).closest('.editable_parent');
			var block = block_parent.find('.editable_block').eq(0);
			var newContent = block.find('input').eq(0).val();
			block.html(newContent);

			var form_action = $(this).closest('a').attr("href");
			//console.log("Form Action: " + form_action);
			$.ajax({
				url: form_action + "?session=" + session,
				type: "POST",
				dataType: "json",
				data: {is_js: "1",edit_item: newContent}
			});
		});
	});

	$(".edit_fields").click(function(event)
	{
		event.preventDefault();
		//console.log("clicked .edit_fields");
		var block_parent = $(this).closest('.editable_parent');
		//console.log(block_parent);
		block_parent.find('.hide_on_edit').hide();
		block_parent.find('.show_on_edit').show();
		block_parent.find('.editable_block').show();
		block_parent.find('.editable_block').each(function(){
			var field_name = this.getAttribute("data-field");
			var field_type = this.getAttribute("data-type");
			if(field_type=="list")
			{
				var field_value = this.getAttribute("data-value");
				if(field_name in form_vars) var it = form_vars[field_name];
				else var it = ['No','Yes'];
				var itLen = it.length;
				var out = "";
				//console.log("Field Name '" + field_name + "'")
				//console.log("Field Type",field_type)
				//console.log("Field Value '" + field_value + "'")
				for (var i = 0; i < itLen; i++) {
					if(field_value == i || field_value == it[i]) {
						sel = "selected ";
						this.classList.remove(field_name + '_' + it[i]);
						this.innerHTML = "";
					} else sel = "";
					out += "<option "+sel+"value='"+i+"'>"+it[i]+"</option>";
				}
				this.innerHTML = "<select data-field='"+field_name+"' name='"+field_name+"'>"+out+"</select>";
			}
			else if(field_type=="hidden") {}
			else this.innerHTML = "<input name='"+field_name+"' value='"+this.textContent+"' type='text'/>";
		});

		// Remove any handlers already attached to the submitter
		$(".submit_edit").unbind("click");

		$(".submit_edit").click(function(event)
		{
			event.preventDefault();
			//console.log("running .submit_edit event");
			var out_data = {is_js: "1"}
			var block_parent = $(this).closest('.editable_parent');
			var block = block_parent.find('.editable_block').each(function(){
				var field_name = this.getAttribute("data-field");
				var field_type = this.getAttribute("data-type");
				if(field_type=="list") {
					var newContent = $(this).find('select :selected').text();
					this.classList.add(field_name + '_' + newContent);
					this.innerHTML = "";
				} else if(field_type=="hidden") {
					var newContent = $(this).val();
				} else {
					var newContent = $(this).find('input').eq(0).val();
					this.innerHTML = newContent;
				}
				this.setAttribute("data-value",newContent);
				out_data[field_name] = newContent;
			});

			var form_action = $(this).closest('a').attr("href");
			//console.log("Form Action: " + form_action);
			//console.log(out_data);
			$.ajax({ url: form_action + "?session=" + session, type:"POST", dataType:"json", data: out_data });
			block_parent.find('.hide_on_edit').show();
			block_parent.find('.show_on_edit').hide();
		});
	});

	$(".ip_item").each(function(){
		var ip = this.textContent;
		if(ip.length > 10){
			this.innerHTML = "Show IP";
			this.onclick = function(event){
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
		var menu_alerts = $(this).parent();
		if(menu_alerts.hasClass("selectedAlert")) {
			event.stopPropagation();
			menu_alerts.removeClass("selectedAlert");
			$("#back").removeClass("alertActive");
		}
	});

	var alert_menu_list = document.getElementsByClassName("menu_alerts");
	for(var i = 0; i < alert_menu_list.length; i++) {
		load_alerts(alert_menu_list[i]);
	}

	$(".menu_alerts").click(function(event) {
		event.stopPropagation();
		if($(this).hasClass("selectedAlert")) return;
		if(!conn) load_alerts(this);
		this.className += " selectedAlert";
		document.getElementById("back").className += " alertActive"
	});

	$("input,textarea,select,option").keyup(function(event){
		event.stopPropagation();
	})

	this.onkeyup = function(event) {
		if(event.which == 37) this.querySelectorAll("#prevFloat a")[0].click();
		if(event.which == 39) this.querySelectorAll("#nextFloat a")[0].click();
	};
});
