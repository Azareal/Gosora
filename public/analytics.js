/*addHook(() => {

})*/

function buildStatsChart(rawLabels, seriesData, timeRange) {
	let labels = [];
	if(timeRange=="one-month") {
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
	seriesData = seriesData.reverse();

	Chartist.Line('.ct_chart', {
		labels: labels,
		series: [seriesData],
	}, {
		height: '250px',
	});
}