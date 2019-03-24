"use strict";

(() => {
addInitHook("end_init", () => {
	$("#dash_username input").click(function(){
		$("#dash_username button").show();
	});
});
})();