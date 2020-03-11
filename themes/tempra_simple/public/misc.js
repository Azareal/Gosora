(() => {
addInitHook("end_init", () => {
	// TODO: Run this when the image is loaded rather than when the document is ready?
	$(".topic_list img").each(function(){
		let aspectRatio = this.naturalHeight / this.naturalWidth;
		console.log("aspectRatio",aspectRatio);
		console.log("height",this.naturalHeight);
		console.log("width",this.naturalWidth);

		$(this).css({ height: aspectRatio * this.width });
	});
});
})()