function post_link(event)
{
	event.preventDefault();
	var form_action = $(event.target).closest('a').attr("href");
	console.log("Form Action: " + form_action);
	$.ajax({
		url: form_action,
		type: "POST",
		dataType: "json",
		data: {js: "1"}
	});
}

$(document).ready(function(){
	$(".open_edit").click(function(event){
		console.log("Clicked on edit");
		event.preventDefault();
		$(".hide_on_edit").hide();
		$(".show_on_edit").show();
	});
	
	$(".submit_edit").click(function(event){
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
		console.log("New Topic Name: " + topic_name_input);
		console.log("New Topic Status: " + topic_status_input);
		console.log("Form Action: " + form_action);
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
	
	$(".post_link").click(post_link);
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
			console.log("Form Action: " + form_action);
			$.ajax({
				url: form_action,
				type: "POST",
				dataType: "json",
				data: {is_js: "1",edit_item: newContent}
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
			console.log("Form Action: " + form_action);
			$.ajax({
				url: form_action + "?session=" + session,
				type: "POST",
				dataType: "json",
				data: {is_js: "1",edit_item: newContent}
			});
		});
	});
	
	$(this).find(".ip_item").each(function(){
		var ip = $(this).text();
		//var ip_width = $(this).width();
		console.log("IP: " + ip);
		if(ip.length > 10){
			$(this).html("Show IP");
			$(this).click(function(event){
				event.preventDefault();
				$(this).text(ip);/*.animate({width: ip.width},{duration: 1000, easing: 'easeOutBounce'});*/
			});
		}
	});
});