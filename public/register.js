(() => {
	addInitHook("end_init", () => {
	fetch("/api/watches/")
	.then(response => {
		if(response.status!==200) {
			console.log("error");
			console.log("response:", response);
			return;
		}
		response.text().then(data => eval(data));
	})
	.catch(err => console.log("err:", err));
	});
})();