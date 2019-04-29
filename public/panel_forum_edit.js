(() => {
	addInitHook("end_init", () => {
		formVars = {'perm_preset': ['can_moderate','can_post','read_only','no_access','default','custom']};
	});
})();