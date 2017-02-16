var form_vars = {};

function post_link(event)
{
	event.preventDefault();
	var form_action = $(event.target).closest('a').attr("href");
	//console.log("Form Action: " + form_action);
	$.ajax({ url: form_action, type: "POST", dataType: "json", data: { js: "1" } });
}

$(document).ready(function(){
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
			data: {
				topic_name: topic_name_input,
				topic_status: topic_status_input,
				topic_content: topic_content_input,
				topic_js: 1
			},
			type: "POST",
			dataType: "json"
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
		var block_parent = $(this).closest('.editable_parent');
		block_parent.find('.hide_on_edit').hide();
		block_parent.find('.editable_block').show();
		block_parent.find('.editable_block').each(function(){
			var field_name = this.getAttribute("data-field");
			var field_type = this.getAttribute("data-type");
			if(field_type=="list") {
				var field_value = this.getAttribute("data-value");
				if(field_name in form_vars) var it = form_vars[field_name];
				else var it = ['No','Yes'];
				var itLen = it.length;
				var out = "";
				for (var i = 0; i < itLen; i++){
					if(field_value==i) sel = "selected ";
					else sel = "";
					out += "<option "+sel+"value='"+i+"'>"+it[i]+"</option>";
				}
				this.innerHTML = "<select data-field='"+field_name+"' name='"+field_name+"'>" + out + "</select>";
			}
			else this.innerHTML = "<input name='"+field_name+"' value='" + this.textContent + "' type='text'/>";
		});
		block_parent.find('.show_on_edit').eq(0).show();
		
		$(".submit_edit").click(function(event)
		{
			event.preventDefault();
			var out_data = {is_js: "1"}
			var block_parent = $(this).closest('.editable_parent');
			var block = block_parent.find('.editable_block').each(function(){
				var field_name = this.getAttribute("data-field");
				var field_type = this.getAttribute("data-type");
				if(field_type == "list") var newContent = $(this).find('select :selected').text();
				else var newContent = $(this).find('input').eq(0).val();
				
				this.innerHTML = newContent;
				out_data[field_name] = newContent
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
		//console.log("IP: " + ip);
		if(ip.length > 10){
			this.innerHTML = "Show IP";
			this.onclick = function(event){
				event.preventDefault();
				this.textContent = ip;
			};
		}
	});
	
	this.onkeyup = function(event){
		if(event.which == 37) this.querySelectorAll("#prevFloat a")[0].click();
		if(event.which == 39) this.querySelectorAll("#nextFloat a")[0].click();
	};
});