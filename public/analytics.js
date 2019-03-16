/*addHook(() => {

})*/

// TODO: Fully localise this
// TODO: Load rawLabels and seriesData dynamically rather than potentially fiddling with nonces for the CSP?
function buildStatsChart(rawLabels, seriesData, timeRange, legendNames) {
	console.log("buildStatsChart");
	let labels = [];
	let aphrases = phraseBox["analytics"];
	if(timeRange=="one-year") {
		labels = [aphrases["analytics.now"],"1" + aphrases["analytics.months_short"]];
		for(let i = 2; i < 12; i++) {
			let label = i + aphrases["analytics.months_short"];
			labels.push(label);
		}
	} else if(timeRange=="three-months") {
		labels = [aphrases["analytics.now"],"3" + aphrases["analytics.days_short"]]
		for(let i = 6; i < 90; i = i + 3) {
			let label = i + aphrases["analytics.days_short"];
			labels.push(label);
		}
	} else if(timeRange=="one-month") {
		labels = [aphrases["analytics.now"],"1" + aphrases["analytics.days_short"]];
		for(let i = 2; i < 30; i++) {
			let label = i + aphrases["analytics.days_short"];
			labels.push(label);
		}
	} else if(timeRange=="one-week") {
		labels = [aphrases["analytics.now"]];
		for(let i = 2; i < 14; i++) {
			if (i%2==0) labels.push("");
			else labels.push(Math.floor(i/2) + aphrases["analytics.days"]);
		}
	} else {
		for(const i in rawLabels) {
			let date = new Date(rawLabels[i]*1000);
			console.log("date: ", date);
			let minutes = "0" + date.getMinutes();
			let label = date.getHours() + ":" + minutes.substr(-2);
			console.log("label:", label);
			labels.push(label);
		}
	}
	labels = labels.reverse()
	for(let i = 0; i < seriesData.length;i++) {
		seriesData[i] = seriesData[i].reverse();
	}

	let config = {
		height: '250px',
	};
	if(legendNames.length > 0) config.plugins = [
		Chartist.plugins.legend({
			legendNames: legendNames,
		})
    ];
	Chartist.Line('.ct_chart', {
		labels: labels,
		series: seriesData,
	}, config);
}