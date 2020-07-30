'use strict';
var me={};
var phraseBox={};
if(tmplInits===undefined) var tmplInits={};
var tmplPhrases=[]; // [key] array of phrases indexed by order of use
var hooks={};
var ranInitHooks={}
var log=console.log;
var pre="/s/";

function runHook(name,...args) {
	if(!(name in hooks)) {
		log("Couldn't find hook "+name);
		return;
	}
	log("Running hook "+name);

	let hook = hooks[name];
	let o;
	for(const index in hook) o = hook[index](...args);
	return o;
}
function addHook(name,h) {
	log("Add hook "+name);
	if(hooks[name]===undefined) hooks[name]=[];
	hooks[name].push(h);
}

// InitHooks are slightly special, as if they are run, then any adds after the initial run will run immediately, this is to deal with the async nature of script loads
function runInitHook(name,...args) {
	ranInitHooks[name]=true;
	return runHook(name,...args);
}
function addInitHook(name,h) {
	addHook(name,h);
	if(name in ranInitHooks) {
		log("Delay running "+name);
		h();
	}
}

// Temporary hack for templates
function len(d) {return d.length}

function asyncGetScript(src) {
	return new Promise((resolve,reject) => {
		let script = document.createElement('script');
		script.async = true;

		const onloadHandler = (e,isAbort) => {
			if (isAbort||!script.readyState||/loaded|complete/.test(script.readyState)) {
				script.onload=null;
				script.onreadystatechange=null;
				script=undefined;

				isAbort ? reject(e) : resolve();
			}
		}
		script.onerror = e => {
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
	src = pre+src;
	return new Promise((resolve,reject) => {
		let ss = src.replace(pre,"");
		try {
			let ssp = ss.charAt(0).toUpperCase() + ss.slice(1)
			log("ssp",ssp)
			if(window[ssp]) {
				resolve();
				return;
			}
		} catch(e) {}
		
		log("src",src)
		let script = document.querySelectorAll(`[src^="${src}"]`)[0];
		log("script",script);
		if(script===undefined) {
			reject("no script found");
			return;
		}

		const onloadHandler = e => {
			script.onload=null;
			script.onreadystatechange=null;
			resolve();
		}
		script.onerror = e => {
			reject(e);
		};
		script.onload = onloadHandler;
		script.onreadystatechange = onloadHandler;
	});
}

function notifyOnScriptW(name,complete,success) {
	notifyOnScript(name)
		.then(() => {
			log(`Loaded ${name}.js`);
			complete();
			if(success!==undefined) success();
		}).catch(e => {
			log("Unable to get "+name,e);
			console.trace();
			complete(e);
		});
}

// TODO: Send data at load time so we don't have to rely on a fallback template here
function loadScript(name,h,fail) {
	let fname = name;
	let value = "; "+document.cookie;
	let parts = value.split("; current_theme=");
	if(parts.length==2) fname += "_"+parts.pop().split(";").shift();
	
	let url = pre+fname+".js"
	let iurl = pre+name+".js"
	asyncGetScript(url)
		.then(h).catch(e => {
			log("Unable to get "+url,e);
			if(fname!=name) {
				asyncGetScript(iurl)
					.then(h).catch(e => {
						log("Unable to get "+iurl,e);
						console.trace();
					});
			}
			console.trace();
			fail(e);
		});
}

function RelativeTime(date) {return date}

function initPhrases(member,acp=false) {
	log("initPhrases")
	log("tmlInits",tmplInits)
	let e = "";
	if(member && !acp) e=",status,topic_list,topic";
	else if(acp) e=",analytics,panel"; // TODO: Request phrases for just one section of the acp?
	else e=",status,topic_list";
	fetchPhrases("alerts,paginator"+e) // TODO: Break this up?
}

function fetchPhrases(plist) {
	fetch("/api/phrases/?q="+plist,{cache:"no-cache"})
		.then(r => r.json())
		.then(d => {
			log("loaded phrase endpoint data",d);
			Object.keys(tmplInits).forEach(key => {
				let phrases=[];
				let tmplInit = tmplInits[key];
				for(let phraseName of tmplInit) phrases.push(d[phraseName]);
				log("Adding phrases for "+key,phrases);
				tmplPhrases[key] = phrases;
			});

			let prefixes={};
			Object.keys(d).forEach(key => {
				let prefix = key.split(".")[0];
				if(prefixes[prefix]===undefined) prefixes[prefix]={};
				prefixes[prefix][key] = d[key];
			});
			Object.keys(prefixes).forEach(prefix => {
				log(`adding phrase prefix ${prefix} to box`);
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
	let q = f => {
		toLoad--;
		if(toLoad===0) initPhrases(member,acp);
		if(f) throw("tmpl func not found");
	};
	let l = (n,h) => notifyOnScriptW("tmpl_"+n,h);

	if(!acp) {
		toLoad += 2;
		if(member) {
			toLoad += 3;
			l("topic_c_edit_post", () => q(!Tmpl_topic_c_edit_post));
			l("topic_c_attach_item", () => q(!Tmpl_topic_c_attach_item));
			l("topic_c_poll_input", () => q(!Tmpl_topic_c_poll_input));
		}
		l("topics_topic", () => q(!Tmpl_topics_topic));
		l("paginator", () => q(!Tmpl_paginator));
	}
	l("notice", () => q(!Tmpl_notice));

	if(member) {
		fetch("/api/me/")
		.then(r => r.json())
		.then(d => {
			log("me",d);
			me=d;
			pre=d.StaticPrefix;
			runInitHook("pre_init");
		});
	} else {
		me={User:{ID:0,S:""},Site:{"MaxReqSize":0}};
		runInitHook("pre_init");
	}
})()