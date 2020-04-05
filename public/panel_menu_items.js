(() => {
	addInitHook("end_init", () => {

// TODO: Move this into a JS file to reduce the number of possible problems
var menuItems = {};
let items = document.getElementsByClassName("panel_menu_item");
for(let i=0; item=items[i]; i++) menuItems[i] = item.getAttribute("data-miid");

Sortable.create(document.getElementById("panel_menu_item_holder"), {
	sort: true,
	onEnd: evt => {
		console.log("pre menuItems",menuItems)
		console.log("evt",evt)
		let oldMiid = menuItems[evt.newIndex];
		menuItems[evt.oldIndex] = oldMiid;
		let newMiid = evt.item.getAttribute("data-miid");
		console.log("newMiid",newMiid);
		menuItems[evt.newIndex] = newMiid;
		console.log("post menuItems",menuItems);
	}
});

document.getElementById("panel_menu_items_order_button").addEventListener("click", () => {
	let req = new XMLHttpRequest();
	if(!req) {
		console.log("Failed to create request");
		return false;
	}
	req.onreadystatechange = () => {
		try {
			if(req.readyState!==XMLHttpRequest.DONE) return;
			// TODO: Signal the error with a notice
			if(req.status===200) {
				let resp = JSON.parse(req.responseText);
				console.log("resp",resp);
				// TODO: Should we move other notices into TmplPhrases like this one?
				pushNotice(phraseBox["panel"]["panel.themes_menus_items_order_updated"]);
				if(resp.success==1) return;
			}
		} catch(e) {
			console.error("e",e)
		}
		console.trace();
	}
	// ? - Is encodeURIComponent the right function for this?
	let spl = document.location.pathname.split("/");
	req.open("POST","/panel/themes/menus/item/order/edit/submit/"+parseInt(spl[spl.length-1],10)+"?s="+encodeURIComponent(me.User.S));
	req.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
	let items = "";
	for(let i=0; item=menuItems[i];i++) items += item+",";
	if(items.length > 0) items = items.slice(0,-1);
	req.send("js=1&amp;items={"+items+"}");
});

});
})()