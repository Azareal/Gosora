function handle_profile_hashbit() {
	var hash_class = "";
	switch(window.location.hash.substr(1)) {
		case "ban_user":
			hash_class = "ban_user_hash";
			break;
		default:
			console.log("Unknown hashbit");
			return;
	}
	$(".hash_hide").hide();
	$("." + hash_class).show();
}

(() => {
addInitHook("end_init", () => {
	if(window.location.hash) handle_profile_hashbit();
	window.addEventListener("hashchange", handle_profile_hashbit, false);
});
})();