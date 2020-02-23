function memStuff(window, document, Chartist) {
	'use strict';

	Chartist.plugins = Chartist.plugins || {};
	Chartist.plugins.byteUnits = function(options) {
	options = Chartist.extend({}, {}, options);

    return function byteUnits(chart) {
    	if(!chart instanceof Chartist.Line) return;
			
		chart.on('created', function() {
			console.log("running created")
			const vbits = document.getElementsByClassName("ct-vertical");
			if(vbits==null) return;

			let tbits = [];
			for(let i = 0; i < vbits.length; i++) {
				tbits[i] = vbits[i].innerHTML;
			}
			console.log("tbits:",tbits);
			
			const calc = (places) => {
				if(places==3) return;
			
				const matcher = vbits[0].innerHTML;
				let allMatch = true;
       			for(let i = 0; i < tbits.length; i++) {
					let val = convertByteUnit(tbits[i], places);
					if(val!=matcher) allMatch = false;
					vbits[i].innerHTML = val;
				}
					
				if(allMatch) calc(places + 1);
			}
			calc(0);
       });
    };
  };
}

function perfStuff(window, document, Chartist) {
	'use strict';

	Chartist.plugins = Chartist.plugins || {};
	Chartist.plugins.perfUnits = function(options) {
	options = Chartist.extend({}, {}, options);

    return function perfUnits(chart) {
    	if(!chart instanceof Chartist.Line) return;
			
		chart.on('created', function() {
			console.log("running created")
			const vbits = document.getElementsByClassName("ct-vertical");
			if(vbits==null) return;

			let tbits = [];
			for(let i = 0; i < vbits.length; i++) {
				tbits[i] = vbits[i].innerHTML;
			}
			console.log("tbits:",tbits);
			
			const calc = (places) => {
				if(places==3) return;
			
				const matcher = vbits[0].innerHTML;
				let allMatch = true;
       			for(let i = 0; i < tbits.length; i++) {
					let val = convertPerfUnit(tbits[i], places);
					if(val!=matcher) allMatch = false;
					vbits[i].innerHTML = val;
				}
					
				if(allMatch) calc(places + 1);
			}
			calc(0);
       });
    };
  };
}

const Kilobyte = 1024;
const Megabyte = Kilobyte * 1024;
const Gigabyte = Megabyte * 1024;
const Terabyte = Gigabyte * 1024;
const Petabyte = Terabyte * 1024;

function convertByteUnit(bytes, places = 0) {
	let out;
	if(bytes >= Petabyte) out = [bytes / Petabyte, "PB"];
	else if(bytes >= Terabyte) out = [bytes / Terabyte, "TB"];
	else if(bytes >= Gigabyte) out = [bytes / Gigabyte, "GB"];
	else if(bytes >= Megabyte) out = [bytes / Megabyte, "MB"];
	else if(bytes >= Kilobyte) out = [bytes / Kilobyte, "KB"];
	else out = [bytes,"b"];

	if(places==0) return Math.ceil(out[0]) + out[1];
	else {
		let ex = Math.pow(10, places);
		return (Math.round(out[0], ex) / ex) + out[1];
	}
}

let ms = 1000;
let sec = ms * 1000;
let min = sec * 60;
let hour = min * 60;
let day = hour * 24;
function convertPerfUnit(quan, places = 0) {
	let out;
	if(quan >= day) out = [quan / day, "d"];
	else if(quan >= hour) out = [quan / hour, "h"];
	else if(quan >= min) out = [quan / min, "m"];
	else if(quan >= sec) out = [quan / sec, "s"];
	else if(quan >= ms) out = [quan / ms, "ms"];
	else out = [quan,"μs"];

	if(places==0) return Math.ceil(out[0]) + out[1];
	else {
		let ex = Math.pow(10, places);
		return (Math.round(out[0], ex) / ex) + out[1];
	}
}

// TODO: Fully localise this
// TODO: Load rawLabels and seriesData dynamically rather than potentially fiddling with nonces for the CSP?
function buildStatsChart(rawLabels, seriesData, timeRange, legendNames, typ=0) {
	console.log("buildStatsChart");
	console.log("seriesData:",seriesData);
	let labels = [];
	let aphrases = phraseBox["analytics"];
	if(timeRange=="one-year") {
		labels = [aphrases["analytics.now"],"1" + aphrases["analytics.months_short"]];
		for(let i = 2; i < 12; i++) {
			labels.push(i + aphrases["analytics.months_short"]);
		}
	} else if(timeRange=="three-months") {
		labels = [aphrases["analytics.now"],"3" + aphrases["analytics.days_short"]]
		for(let i = 6; i < 90; i = i + 3) {
			if (i%2==0) labels.push("");
			else labels.push(i + aphrases["analytics.days_short"]);
		}
	} else if(timeRange=="one-month") {
		labels = [aphrases["analytics.now"],"1" + aphrases["analytics.days_short"]];
		for(let i = 2; i < 30; i++) {
			if (i%2==0) labels.push("");
			else labels.push(i + aphrases["analytics.days_short"]);
		}
	} else if(timeRange=="one-week") {
		labels = [aphrases["analytics.now"]];
		for(let i = 2; i < 14; i++) {
			if (i%2==0) labels.push("");
			else labels.push(Math.floor(i/2) + aphrases["analytics.days"]);
		}
	} else if (timeRange=="two-days" || timeRange == "one-day" || timeRange == "twelve-hours") {
		for(const i in rawLabels) {
			if (i%2==0) {
				labels.push("");
				continue;
			}
			let date = new Date(rawLabels[i]*1000);
			console.log("date:", date);
			let minutes = "0" + date.getMinutes();
			let label = date.getHours() + ":" + minutes.substr(-2);
			console.log("label:", label);
			labels.push(label);
		}
	} else {
		for(const i in rawLabels) {
			let date = new Date(rawLabels[i]*1000);
			console.log("date:", date);
			let minutes = "0" + date.getMinutes();
			let label = date.getHours() + ":" + minutes.substr(-2);
			console.log("label:", label);
			labels.push(label);
		}
	}
	labels = labels.reverse()
	for(let i = 0; i < seriesData.length; i++) {
		seriesData[i] = seriesData[i].reverse();
	}

	let config = {height: '250px', plugins:[]};
	if(legendNames.length > 0) config.plugins = [
		Chartist.plugins.legend({legendNames: legendNames})
		];
	if(typ==1) config.plugins.push(Chartist.plugins.byteUnits());
	else if(typ==2) config.plugins.push(Chartist.plugins.perfUnits());
	Chartist.Line('.ct_chart', {
		labels: labels,
		series: seriesData,
	}, config);
}

runInitHook("analytics_loaded");