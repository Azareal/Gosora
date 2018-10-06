"use strict";

(() => {
	addInitHook("after_update_alert_list", () => {
		if(alertCount==0) {
			$(".alerts").html("No new alerts");
			$(".user_box").removeClass("has_alerts");
		} else {
			$(".alerts").html(alertCount + " new alerts");
			$(".user_box").addClass("has_alerts");
		}
	})
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