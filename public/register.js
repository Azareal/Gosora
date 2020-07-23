(() => {
	addInitHook("end_init", () => {
	fetch("/api/watches/")
	.then(resp => {
		if(resp.status!==200) {
			log("err");
			log("resp",resp);
			return;
		}
		resp.text().then(d => eval(d));
	})
	.catch(e => log("e",e));
	});
})()