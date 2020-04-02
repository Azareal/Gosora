(() => {
	addInitHook("end_init", () => {
	fetch("/api/watches/")
	.then(resp => {
		if(resp.status!==200) {
			console.log("err");
			console.log("resp",resp);
			return;
		}
		resp.text().then(dat => eval(dat));
	})
	.catch(e => console.log("e",e));
	});
})()