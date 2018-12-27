"use strict";

(() => {
	addInitHook("after_update_alert_list", (alertCount) => {
		console.log("misc.js");
		console.log("alertCount:",alertCount);
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
})();

$(document).ready(() => {
	$(".alerts").click((event) => {
		event.stopPropagation();
		var alerts = $(".menu_alerts")[0];
		if($(alerts).hasClass("selectedAlert")) return;
		if(!conn) loadAlerts(alerts);
		alerts.className += " selectedAlert";
		document.getElementById("back").className += " alertActive"
	});
});