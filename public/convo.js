(() => {
	addInitHook("end_init", () => {
		$(".create_convo_link").click((event) => {
			event.preventDefault();
			$(".convo_create_form").removeClass("auto_hide");
		});
		$(".convo_create_form .close_form").click((event) => {
			event.preventDefault();
			$(".convo_create_form").addClass("auto_hide");
		});
	});
})();