$(document).ready(function(){
	// Is there we way we can append instead? Maybe, an editor plugin?
	attachItemCallback = function(attachItem) {
		let currentContent = $('#input_content').trumbowyg('html');
		$('#input_content').trumbowyg('html', currentContent);
	}
	
	$(".topic_name_row").click(function(){
		$(".topic_create_form").addClass("selectedInput");
	});
	//$.trumbowyg.svgPath = false;
	$('#input_content').trumbowyg({
		btns: [['viewHTML'],['undo','redo'],['formatting'],['strong','em','del'],['link'],['insertImage'],['unorderedList','orderedList'],['removeformat']],
		//hideButtonTexts: true
	});
});