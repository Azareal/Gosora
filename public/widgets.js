"use strict";

$(document).ready(() => {
	let clickHandle = function(event){
		console.log("in clickHandle")
		event.preventDefault();
		let eparent = $(this).closest(".editable_parent");
		eparent.find(".hide_on_block_edit").addClass("edit_opened");
		eparent.find(".show_on_block_edit").addClass("edit_opened");
		eparent.addClass("in_edit");

		eparent.find(".widget_save").click(() => {
			eparent.find(".hide_on_block_edit").removeClass("edit_opened");
			eparent.find(".show_on_block_edit").removeClass("edit_opened");
			eparent.removeClass("in_edit");
		});

		eparent.find(".widget_delete").click(function(event) {
			event.preventDefault();
			eparent.remove();
			let formData = new URLSearchParams();
			formData.append("session",me.User.Session);
			let req = new XMLHttpRequest();
			let target = this.closest("a").getAttribute("href");
			req.open("POST",target,true);
			req.send(formData);
		});
	};

	$(".widget_item a").click(clickHandle);

	let changeHandle = function(event){
		let wtype = this.options[this.selectedIndex].value;
		let typeBlock = this.closest(".widget_edit").querySelector(".wtypes");
		typeBlock.className = "wtypes wtype_"+wtype;
	};

	$(".wtype_sel").change(changeHandle);

	$(".widget_new a").click(function(event){
		console.log("clicked widget_new a")
		let widgetList = this.closest(".panel_widgets");
		let widgetNew = this.closest(".widget_new");
		let widgetTmpl = document.getElementById("widgetTmpl").querySelector(".widget_item");
		let node = widgetTmpl.cloneNode(true);
		node.querySelector(".wside").value = this.getAttribute("data-dock");
		widgetList.insertBefore(node,widgetNew);
		$(".widget_item a").unbind("click");
		$(".widget_item a").click(clickHandle);
		$(".wtype_sel").unbind("change");
		$(".wtype_sel").change(changeHandle);
	});

	$(".widget_save").click(function(event){
		console.log("in .widget_save")
		event.preventDefault();
		event.stopPropagation();
		let pform = this.closest("form");
		let data = new URLSearchParams();
		for (const pair of new FormData(pform)) data.append(pair[0], pair[1]);
		data.append("session",me.User.Session);
		var req = new XMLHttpRequest();
		req.open("POST", pform.getAttribute("action"));
		req.send(data);
	});
});