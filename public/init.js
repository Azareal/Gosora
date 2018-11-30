'use strict';

var me = {};
var phraseBox = {};
var tmplInits = {};
var tmplPhrases = []; // [key] array of phrases indexed by order of use
var hooks = {
	"pre_iffe": [],
	"pre_init": [],
	"start_init": [],
	"end_init": [],
	"after_phrases":[],
	"after_add_alert":[],
	"after_update_alert_list":[],
};
var ranInitHooks = {}

function runHook(name, ...args) {
	if(!(name in hooks)) {
		console.log("Couldn't find hook '" + name + "'");
		return;
	}
	console.log("Running hook '"+name+"'");

	let hook = hooks[name];
	for (const index in hook) {
		hook[index](...args);
	}
}

function addHook(name, callback) {
	hooks[name].push(callback);
}

// InitHooks are slightly special, as if they are run, then any adds after the initial run will run immediately, this is to deal with the async nature of script loads
function runInitHook(name, ...args) {
	runHook(name,...args);
	ranInitHooks[name] = true;
}

function addInitHook(name, callback) {
	addHook(name, callback);
	if(name in ranInitHooks) {
		callback();
	}
}

// Temporary hack for templates
function len(item) {
	return item.length;
}

const asyncGetScript = (source) => {
	return new Promise((resolve, reject) => {
		let script = document.createElement('script');
		script.async = true;

		const onloadHander = (haha, isAbort) => {
			if (isAbort || !script.readyState || /loaded|complete/.test(script.readyState)) {
				script.onload = null;
				script.onreadystatechange = null;
				script = undefined;

				isAbort ? reject(haha) : resolve();
			}
		}

		script.onerror = (haha) => {
			reject(haha);
		};
		script.onload = onloadHander;
		script.onreadystatechange = onloadHander;
		script.src = source;

		const prior = document.getElementsByTagName('script')[0];
		prior.parentNode.insertBefore(script, prior);
	});
};

function loadScript(name, callback) {
	let url = "/static/"+name
	asyncGetScript(url)
		.then(callback)
		.catch((haha) => {
			console.log("Unable to get script '"+url+"'");
			console.log("haha: ", haha);
			console.trace();
		});
}

function DoNothingButPassBack(item) {
	return item;
}

function fetchPhrases() {
	fetch("/api/phrases/?query=status,topic_list,alerts")
		.then((resp) => resp.json())
		.then((data) => {
			console.log("loaded phrase endpoint data");
			console.log("data:",data);
			Object.keys(tmplInits).forEach((key) => {
				let phrases = [];
				let tmplInit = tmplInits[key];
				for(let phraseName of tmplInit) {
					phrases.push(data[phraseName]);
				}
				console.log("Adding phrases");
				console.log("key:",key);
				console.log("phrases:",phrases);
				tmplPhrases[key] = phrases;
			});

			let prefixes = {};
			Object.keys(data).forEach((key) => {
				let prefix = key.split(".")[0];
				if(prefixes[prefix]===undefined) {
					prefixes[prefix] = {};
				}
				prefixes[prefix][key] = data[key];
			});
			Object.keys(prefixes).forEach((prefix) => {
				console.log("adding phrase prefix '"+prefix+"' to box");
				phraseBox[prefix] = prefixes[prefix];
			});

			runInitHook("after_phrases");
		});
}

(() => {
	runInitHook("pre_iife");
	let loggedIn = document.head.querySelector("[property='x-loggedin']").content;
	if(loggedIn) {
		fetch("/api/me/")
		.then((resp) => resp.json())
		.then((data) => {
			console.log("loaded me endpoint data");
			console.log("data:",data);
			me = data;
			runInitHook("pre_init");
		});
		
		let toLoad = 1;
		loadScript("template_topics_topic.js", () => {
			console.log("Loaded template_topics_topic.js");
			toLoad--;
			if(toLoad===0) fetchPhrases();
		});
	} else {
		me = {User:{ID:0,Session:""},Site:{"MaxRequestSize":0}};
		runInitHook("pre_init");
	}
})();