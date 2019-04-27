(() => {
addInitHook("end_init", () => {

formVars = {
	'forum_active': ['Hide','Show'],
	'forum_preset': ['all','announce','members','staff','admins','archive','custom']
};
var forums = {};
let items = document.getElementsByClassName("panel_forum_item");
for(let i = 0; item = items[i]; i++) forums[i] = item.getAttribute("data-fid");
console.log("forums:",forums);

Sortable.create(document.getElementById("panel_forums"), {
	sort: true,
	onEnd: (evt) => {
		console.log("pre forums: ", forums)
		console.log("evt: ", evt)
		let oldFid = forums[evt.newIndex];
		forums[evt.oldIndex] = oldFid;
		let newFid = evt.item.getAttribute("data-fid");
		console.log("newFid: ", newFid);
		forums[evt.newIndex] = newFid;
		console.log("post forums: ", forums);
	}
});

document.getElementById("panel_forums_order_button").addEventListener("click", () => {
	let req = new XMLHttpRequest();
	if(!req) {
		console.log("Failed to create request");
		return false;
	}
	req.onreadystatechange = () => {
		try {
			if(req.readyState!==XMLHttpRequest.DONE) return;
			// TODO: Signal the error with a notice
			if(req.status!==200) return;
			
			let resp = JSON.parse(req.responseText);
			console.log("resp: ", resp);
			// TODO: Should we move other notices into TmplPhrases like this one?
			pushNotice(phraseBox["panel"]["panel.forums_order_updated"]);
			if(resp.success==1) return;
		} catch(ex) {
			console.error("exception: ", ex)
		}
		console.trace();
	}
	// ? - Is encodeURIComponent the right function for this?
	req.open("POST","/panel/forums/order/edit/submit/?session=" + encodeURIComponent(me.User.Session));
	req.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
	let items = "";
	for(let i = 0; item = forums[i];i++) items += item+",";
	if(items.length > 0) items = items.slice(0,-1);
	req.send("js=1&amp;items={"+items+"}");
});

});
})();