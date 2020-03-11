"use strict";

function noxMenuBind() {
	$(".more_menu").remove();
	$("#main_menu li:not(.menu_hamburger").removeClass("menu_hide");

	let mWidth = $("#main_menu").width();
	let iWidth = 0;
	let lastElem = null;
	$("#main_menu > li:not(.menu_hamburger)").each(function(){
		iWidth += $(this).outerWidth();
		if(iWidth > (mWidth - 100) && (mWidth - 100) > 0) {
			this.classList.add("menu_hide");
			if(lastElem!==null) lastElem.classList.add("menu_hide");
		}
		lastElem = this;
	});
	if(iWidth > (mWidth - 100) && (mWidth - 100) > 0) $(".menu_hamburger").show();
	else $(".menu_hamburger").hide();

	let div = document.createElement('div');
	div.className = "more_menu";
	$("#main_menu > li:not(.menu_hamburger):not(#menu_overview)").each(function(){
		if(!this.classList.contains("menu_hide")) return;
		let cop = this.cloneNode(true);
		cop.classList.remove("menu_hide");
		div.appendChild(cop);
	});
	document.getElementsByClassName("menu_hamburger")[0].appendChild(div);
}

(() => {
	if(window.location.pathname.startsWith("/panel/")) {
		addInitHook("pre_global", () => noAlerts = true);
	}
	
	function moveAlerts() {
		// Move the alerts above the first header
		let colSel = $(".colstack_right .colstack_head:first");
		let colSelAlt = $(".colstack_right .colstack_item:first");
		let colSelAltAlt = $(".colstack_right .coldyn_block:first");
		if(colSel.length > 0) $('.alert').insertBefore(colSel);
		else if (colSelAlt.length > 0) $('.alert').insertBefore(colSelAlt);
		else if (colSelAltAlt.length > 0) $('.alert').insertBefore(colSelAltAlt);
		else $('.alert').insertAfter(".rowhead:first");
	}
	
	addInitHook("after_update_alert_list", alertCount => {
		console.log("misc.js");
		console.log("alertCount",alertCount);
		if(alertCount==0) {
			$(".alerts").html(phraseBox["alerts"]["alerts.no_alerts_short"]);
			$(".user_box").removeClass("has_alerts");
		} else {
			// TODO: Localise this
			$(".alerts").html(alertCount + " new alerts");
			$(".user_box").addClass("has_alerts");
		}
	});
	addHook("open_edit", () => $('.topic_block').addClass("edithead"));
	addHook("close_edit", () => $('.topic_block').removeClass("edithead"));

	addInitHook("end_init", () => {
		$(".alerts").click(ev => {
			ev.stopPropagation();
			var alerts = $(".menu_alerts")[0];
			if($(alerts).hasClass("selectedAlert")) return;
			if(!conn) loadAlerts(alerts);
			alerts.className += " selectedAlert";
			document.getElementById("back").className += " alertActive"
		});

		$(window).resize(() => noxMenuBind());
		noxMenuBind();
		moveAlerts();

		$(".menu_hamburger").click(function() {
			event.stopPropagation();
			let mm = document.getElementsByClassName("more_menu")[0];
			mm.classList.add("more_menu_selected");
			let calc = $(this).offset().left - (mm.offsetWidth / 4);
			mm.style.left = calc+"px";
		});

		$(document).click(() => $(".more_menu").removeClass("more_menu_selected"));
	});

	addInitHook("after_notice", moveAlerts);
})();