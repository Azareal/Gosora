"use strict";

(() => {
	function newElement(etype, eclass) {
		let element = document.createElement(etype);
		element.className = eclass;
		return element;
	}
	
	function moveAlerts() {
		// Move the alerts under the first header
		let colSel = $(".colstack_right .colstack_head:first");
		let colSelAlt = $(".colstack_right .colstack_item:first");
		let colSelAltAlt = $(".colstack_right .coldyn_block:first");
		if(colSel.length > 0) $('.alert').insertAfter(colSel);
		else if (colSelAlt.length > 0) $('.alert').insertBefore(colSelAlt);
		else if (colSelAltAlt.length > 0) $('.alert').insertBefore(colSelAltAlt);
		else $('.alert').insertAfter(".rowhead:first");
	}
	
	addInitHook("end_init", () => {
		let loggedIn = document.head.querySelector("[property='x-mem']")!=null;
		if(loggedIn) {
			if(navigator.userAgent.indexOf("Firefox")!=-1) $.trumbowyg.svgPath = pre+"trumbowyg/ui/icons.svg";
			
			// Is there we way we can append instead? Maybe, an editor plugin?
			attachItemCallback = function(attachItem) {
				let currentContent = $('#input_content').trumbowyg('html');
				$('#input_content').trumbowyg('html',currentContent);
			}
			quoteItemCallback = function() {
				let currentContent = $('#input_content').trumbowyg('html');
				$('#input_content').trumbowyg('html',currentContent);
			}
			
			$(".topic_name_row").click(() => {
				$(".topic_create_form").addClass("selectedInput");
			});

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
			$('#profile_comments_form .topic_reply_form .input_content').trumbowyg({
				btns: [['viewHTML'],['strong','em','del'],['link'],['insertImage'],['removeformat']],
				autogrow: true,
			});
			addHook("edit_item_pre_bind", () => {
				$('.user_content textarea').trumbowyg({
					btns: btnlist,
					autogrow: true,
				});
			});
		}

		// TODO: Refactor this to use `each` less
		$('.button_menu').click(function(){
			log(".button_menu");
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
				if(contents.length > 45) this.innerHTML = contents.substring(0,45) + "...";
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
			log("rowCount",rowCount);
			log("gridElementCount",gridElementCount);
			if(gridElementCount%rowCount != 0) {
				let fillerNodes = (rowCount - (gridElementCount%rowCount));
				log("fillerNodes",fillerNodes);
				for(let i = 0; i < fillerNodes;i++ ) {
					log("added a gridFiller");
					buttonGrid.appendChild(newElement("div","gridFiller"));
				}
			}
			buttonPane.appendChild(buttonGrid);

			document.getElementById("back").appendChild(buttonPane);
		});

		moveAlerts();
	});

	addInitHook("after_notice", moveAlerts);
})()