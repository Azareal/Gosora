(() => {
	addInitHook("end_init", () => {
	fetch("/api/watches/")
	.then(resp => {
		if(resp.status!==200) {
			console.log("err");
			console.log("resp",resp);
			return;
		}
		resp.text().then(data => eval(data));
	})
	.catch(err => console.log("err", err));
	});
})()