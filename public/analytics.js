/*addHook(() => {

})*/

// TODO: Fully localise this
// TODO: Load rawLabels and seriesData dynamically rather than potentially fiddling with nonces for the CSP?
function buildStatsChart(rawLabels, seriesData, timeRange, legendNames) {
	let labels = [];
	if(timeRange=="one-year") {
		labels = ["today","01 months"];
		for(let i = 2; i < 12; i++) {
			let label = "0" + i + " months";
			if(label.length > "01 months".length) label = label.substr(1);
			labels.push(label);
		}
	} else if(timeRange=="three-months") {
		labels = ["today","01 days"];
		for(let i = 2; i < 90; i = i + 3) {
			let label = "0" + i + " days";
			if(label.length > "01 days".length) label = label.substr(1);
			labels.push(label);
		}
	} else if(timeRange=="one-month") {
		labels = ["today","01 days"];
		for(let i = 2; i < 30; i++) {
			let label = "0" + i + " days";
			if(label.length > "01 days".length) label = label.substr(1);
			labels.push(label);
		}
	} else if(timeRange=="one-week") {
		labels = ["today"];
		for(let i = 2; i < 14; i++) {
			if (i%2==0) labels.push("");
			else labels.push(Math.floor(i/2) + " days");
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