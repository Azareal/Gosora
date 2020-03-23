'use strict';
var me={};
var phraseBox={};
if(tmplInits===undefined) var tmplInits={};
var tmplPhrases=[]; // [key] array of phrases indexed by order of use
var hooks={};
var ranInitHooks={}

function runHook(name,...args) {
	if(!(name in hooks)) {
		console.log("Couldn't find hook '"+name+"'");
		return;
	}
	console.log("Running hook '"+name+"'");

	let hook = hooks[name];
	let ret;
	for (const index in hook) ret = hook[index](...args);
	return ret;
}

function addHook(name,h) {
	if(hooks[name]===undefined) hooks[name]=[];
	hooks[name].push(h);
}

// InitHooks are slightly special, as if they are run, then any adds after the initial run will run immediately, this is to deal with the async nature of script loads
function runInitHook(name,...args) {
	let ret = runHook(name,...args);
	ranInitHooks[name] = true;
	return ret;
}

function addInitHook(name,h) {
	addHook(name, h);
	if(name in ranInitHooks) h();
}

// Temporary hack for templates
function len(it) {return it.length;}

function asyncGetScript(src) {
	return new Promise((resolve,reject) => {
		let script = document.createElement('script');
		script.async = true;

		const onloadHandler = (e,isAbort) => {
			if (isAbort || !script.readyState || /loaded|complete/.test(script.readyState)) {
				script.onload = null;
				script.onreadystatechange = null;
				script = undefined;

				isAbort ? reject(e) : resolve();
			}
		}

		script.onerror = (e) => {
			reject(e);
		};
		script.onload = onloadHandler;
		script.onreadystatechange = onloadHandler;
		script.src = src;

		const prior = document.getElementsByTagName('script')[0];
		prior.parentNode.insertBefore(script,prior);
	});
}

function notifyOnScript(src) {
	src = "/s/"+src;
	return new Promise((resolve,reject) => {
		let ss = src.replace("/s/","");
		try {
			let ssp = ss.charAt(0).toUpperCase() + ss.slice(1)
			console.log("ssp",ssp)
			if(window[ssp]) {
				resolve();
				return;
			}
		} catch(e) {}
		
		console.log("src",src)
		let script = document.querySelectorAll('[src^="'+src+'"]')[0];
		console.log("script",script);
		if(script===undefined) {
			reject("no script found");
			return;
		}

		const onloadHandler = (e) => {
			script.onload = null;
			script.onreadystatechange = null;
			resolve();
		}
		script.onerror = (e) => {
			reject(e);
		};
		script.onload = onloadHandler;
		script.onreadystatechange = onloadHandler;
	});
}

function notifyOnScriptW(name,complete,success) {
	notifyOnScript(name)
		.then(() => {
			console.log("Loaded "+name+".js");
			complete();
			if(success!==undefined) success();
		}).catch(e => {
			console.log("Unable to get '"+name+"'",e);
			console.trace();
			complete(e);
		});
}

// TODO: Send data at load time so we don't have to rely on a fallback template here
function loadScript(name,callback,fail) {
	let fname = name;
	let value = "; "+document.cookie;
	let parts = value.split("; current_theme=");
	if(parts.length==2) fname += "_"+parts.pop().split(";").shift();
	
	let url = "/s/"+fname+".js"
	let iurl = "/s/"+name+".js"
	asyncGetScript(url)
		.then(callback)
		.catch(e => {
			console.log("Unable to get '"+url+"'");
			if(fname!=name) {
				asyncGetScript(iurl)
					.then(callback)
					.catch(e => {
						console.log("Unable to get '"+iurl+"'",e);
						console.trace();
					});
			}
			console.log("e",e);
			console.trace();
			fail(e);
		});
}

function RelativeTime(date) {return date}

function initPhrases(member, acp=false) {
	console.log("initPhrases")
	console.log("tmlInits",tmplInits)
	let e = "";
	if(member && !acp) e = ",status,topic_list,topic";
	else if(acp) e = ",analytics,panel"; // TODO: Request phrases for just one section of the cp?
	else e = ",status,topic_list";
	fetchPhrases("alerts,paginator"+e) // TODO: Break this up?
}

function fetchPhrases(plist) {
	fetch("/api/phrases/?q="+plist,{cache:"no-cache"})
		.then(resp => resp.json())
		.then(data => {
			console.log("loaded phrase endpoint data");
			console.log("data",data);
			Object.keys(tmplInits).forEach(key => {
				let phrases=[];
				let tmplInit = tmplInits[key];
				for(let phraseName of tmplInit) phrases.push(data[phraseName]);
				console.log("Adding phrases");
				console.log("key",key);
				console.log("phrases",phrases);
				tmplPhrases[key] = phrases;
			});

			let prefixes = {};
			Object.keys(data).forEach(key => {
				let prefix = key.split(".")[0];
				if(prefixes[prefix]===undefined) prefixes[prefix] = {};
				prefixes[prefix][key] = data[key];
			});
			Object.keys(prefixes).forEach(prefix => {
				console.log("adding phrase prefix '"+prefix+"' to box");
				phraseBox[prefix] = prefixes[prefix];
			});

			runInitHook("after_phrases");
		});
}

(() => {
	runInitHook("pre_iife");
	let member = document.head.querySelector("[property='x-mem']")!=null;
	let acp = window.location.pathname.startsWith("/panel/");

	let toLoad = 1;
	// TODO: Shunt this into member if there aren't any search and filter widgets?
	let q = (f) => {
		toLoad--;
		if(toLoad===0) initPhrases(member,acp);
		if(f) throw("tmpl func not found");
	};

	if(!acp) {
		toLoad += 2;
		if(member) {
			toLoad += 3;
			notifyOnScriptW("tmpl_topic_c_edit_post", () => q(!Tmpl_topic_c_edit_post));
			notifyOnScriptW("tmpl_topic_c_attach_item", () => q(!Tmpl_topic_c_attach_item));
			notifyOnScriptW("tmpl_topic_c_poll_input", () => q(!Tmpl_topic_c_poll_input));
		}
		notifyOnScriptW("tmpl_topics_topic", () => q(!Tmpl_topics_topic));
		notifyOnScriptW("tmpl_paginator", () => q(!Tmpl_paginator));
	}
	notifyOnScriptW("tmpl_notice", () => q(!Tmpl_notice));

	if(member) {
		fetch("/api/me/")
		.then(resp => resp.json())
		.then(data => {
			console.log("me data",data);
			me=data;
			runInitHook("pre_init");
		});
	} else {
		me={User:{ID:0,S:""},Site:{"MaxRequestSize":0}};
		runInitHook("pre_init");
	}
})()