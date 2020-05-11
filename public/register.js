(() => {
	addInitHook("end_init", () => {
	fetch("/api/watches/")
	.then(resp => {
		if(resp.status!==200) {
			console.log("err");
			console.log("resp",resp);
			return;
		}
		resp.text().then(d => eval(d));
	})
	.catch(e => console.log("e",e));
	});
})()