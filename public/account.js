"use strict";

(() => {
addInitHook("end_init", () => {
	$("#dash_username input").click(()=>{
		$("#dash_username button").show();
	});
});
})();