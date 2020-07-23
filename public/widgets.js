"use strict";
$(document).ready(() => {
	let clickHandle = function(ev){
		log("in clickHandle")
		ev.preventDefault();
		let ep = $(this).closest(".editable_parent");
		ep.find(".hide_on_block_edit").addClass("edit_opened");
		ep.find(".show_on_block_edit").addClass("edit_opened");
		ep.addClass("in_edit");

		ep.find(".widget_save").click(() => {
			ep.find(".hide_on_block_edit").removeClass("edit_opened");
			ep.find(".show_on_block_edit").removeClass("edit_opened");
			ep.removeClass("in_edit");
		});

		ep.find(".widget_delete").click(function(ev) {
			ev.preventDefault();
			ep.remove();
			let formData = new URLSearchParams();
			formData.append("s",me.User.S);
			let req = new XMLHttpRequest();
			let target = this.closest("a").getAttribute("href");
			req.open("POST",target,true);
			req.send(formData);
		});
	};

	$(".widget_item a").click(clickHandle);

	let changeHandle = function(ev){
		let wtype = this.options[this.selectedIndex].value;
		let typeBlock = this.closest(".widget_edit").querySelector(".wtypes");
		typeBlock.className = "wtypes wtype_"+wtype;
	};
	$(".wtype_sel").change(changeHandle);

	$(".widget_new a").click(function(ev){
		log("clicked widget_new a")
		let widgetList = this.closest(".panel_widgets");
		let widgetNew = this.closest(".widget_new");
		let widgetTmpl = document.getElementById("widgetTmpl").querySelector(".widget_item");
		let n = widgetTmpl.cloneNode(true);
		n.querySelector(".wside").value = this.getAttribute("data-dock");
		widgetList.insertBefore(n,widgetNew);
		$(".widget_item a").unbind("click");
		$(".widget_item a").click(clickHandle);
		$(".wtype_sel").unbind("change");
		$(".wtype_sel").change(changeHandle);
	});

	$(".widget_save").click(function(ev){
		log("in .widget_save")
		ev.preventDefault();
		ev.stopPropagation();
		let pform = this.closest("form");
		let dat = new URLSearchParams();
		for (const pair of new FormData(pform)) dat.append(pair[0], pair[1]);
		dat.append("s",me.User.S);
		var req = new XMLHttpRequest();
		req.open("POST",pform.getAttribute("action"));
		req.send(dat);
	});
});