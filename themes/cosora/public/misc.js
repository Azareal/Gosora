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

	// TODO: Bind this to the viewport resize event
	var btnlist = [];
	if(document.documentElement.clientWidth > 550) {
		btnlist = [['viewHTML'],['undo','redo'],['formatting'],['strong','em','del'],['link'],['insertImage'],['unorderedList','orderedList'],['removeformat']];
	} else {
		btnlist = [['viewHTML'],['strong','em','del'],['link'],['insertImage'],['unorderedList','orderedList'],['removeformat']];
	}
	
	$('.topic_create_form #input_content').trumbowyg({
		btns: btnlist,
	});
	$('.topic_reply_form #input_content').trumbowyg({
		btns: btnlist,
		autogrow: true,
	});

	// TODO: Refactor this to use `each` less
	$('.button_menu').click(function(){
		console.log(".button_menu");
		
		// The outer container
		let buttonPane = newElement("div","button_menu_pane");
		let postItem = $(this).parents('.post_item');

		// Create the userinfo row in the pane
		let userInfo = newElement("div","userinfo");
		postItem.find('.avatar_item').each(function(){
			userInfo.appendChild(this);
		});

		let userText = newElement("div","userText");
		postItem.find('.userinfo:not(.avatar_item)').children().each(function(){
			userText.appendChild(this);
		});
		userInfo.appendChild(userText);
		buttonPane.appendChild(userInfo);

		// Copy a short preview of the post contents into the pane
		postItem.find('.user_content').each(function(){
			// TODO: Truncate an excessive number of lines to 5 or so
			let contents = this.innerHTML;
			if(contents.length > 45) {
				this.innerHTML = contents.substring(0,45) + "...";
			}
			buttonPane.appendChild(this);
		});

		// Copy the buttons from the post to the pane
		let buttonGrid = newElement("div","buttonGrid");
		let gridElementCount = 0;
		$(this).parent().children('a:not(.button_menu)').each(function(){
			buttonGrid.appendChild(this);
			gridElementCount++;
		});
		

		// Fill in the placeholder grid nodes
		let rowCount = 4;
		console.log("rowCount: ",rowCount);
		console.log("gridElementCount: ",gridElementCount);
		if(gridElementCount%rowCount != 0) {
			let fillerNodes = (rowCount - (gridElementCount%rowCount));
			console.log("fillerNodes: ",fillerNodes);
			for(let i = 0; i < fillerNodes;i++ ) {
				console.log("added a gridFiller");
				buttonGrid.appendChild(newElement("div","gridFiller"));
			}
		}
		buttonPane.appendChild(buttonGrid);

		document.getElementById("back").appendChild(buttonPane);
	});
});

function newElement(etype, eclass) {
	let element = document.createElement(etype);
	element.className = eclass;
	return element;
}