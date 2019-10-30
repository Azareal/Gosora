// TODO: Push ImageFileExts to the client from the server in some sort of gen.js?
var imageExts = ["png", "jpg", "jpe","jpeg","jif","jfi","jfif", "svg", "bmp", "gif", "tiff","tif", "webp"];

(() => {
	addInitHook("almost_end_init", () => {
	function copyToClipboard(str) {
		const el = document.createElement('textarea');
		el.value = str;
		el.setAttribute('readonly', '');
		el.style.position = 'absolute';
		el.style.left = '-9999px';
		document.body.appendChild(el);
		el.select();
		document.execCommand('copy');
		document.body.removeChild(el);
	}

	// TODO: Surely, there's a prettier and more elegant way of doing this?
	function getExt(filename) {
		if(!filename.indexOf('.' > -1)) throw("This file doesn't have an extension");
		return filename.split('.').pop();
	}
		
	function uploadFileHandler(fileList, maxFiles = 5, step1 = () => {}, step2 = () => {}) {
		let files = [];
		for(var i = 0; i < fileList.length && i < 5; i++) files[i] = fileList[i];
	
		let totalSize = 0;
		for(let i = 0; i < files.length; i++) {
			console.log("files[" + i + "]",files[i]);
			totalSize += files[i]["size"];
		}
		if(totalSize > me.Site.MaxRequestSize) {
			throw("You can't upload this much at once, max: " + me.Site.MaxRequestSize);
		}
	
		for(let i = 0; i < files.length; i++) {
			let filename = files[i]["name"];
			let f = (e) => {
				step1(e,filename)
					
				let reader = new FileReader();
				reader.onload = (e2) => {
					crypto.subtle.digest('SHA-256',e2.target.result)
						.then((hash) => {
							const hashArray = Array.from(new Uint8Array(hash))
							return hashArray.map(b => ('00' + b.toString(16)).slice(-2)).join('')
						}).then(hash => step2(e,hash,filename));
				}
				reader.readAsArrayBuffer(files[i]);
			};
				
			let ext = getExt(filename);
			// TODO: Push ImageFileExts to the client from the server in some sort of gen.js?
			let isImage = imageExts.includes(ext);
			if(isImage) {
				let reader = new FileReader();
				reader.onload = f;
				reader.readAsDataURL(files[i]);
			} else f(null);
		}
	}

	// Attachment Manager
	function uploadAttachHandler2() {
		let post = this.closest(".post_item");
		let fileDock = this.closest(".attach_edit_bay");
		try {
			uploadFileHandler(this.files, 5, () => {},
			(e, hash, filename) => {
				console.log("hash",hash);

				let formData = new FormData();
				formData.append("s",me.User.S);
				for(let i = 0; i < this.files.length; i++) formData.append("upload_files",this.files[i]);
				bindAttachManager();

				let req = new XMLHttpRequest();
				req.addEventListener("load", () => {
					let data = JSON.parse(req.responseText);
					//console.log("rdata:",data);
					let fileItem = document.createElement("div");
					let ext = getExt(filename);
					// TODO: Push ImageFileExts to the client from the server in some sort of gen.js?
					let isImage = imageExts.includes(ext);
					let c = "";
					if(isImage) c = " attach_image_holder"
					fileItem.className = "attach_item attach_item_item" + c;
					fileItem.innerHTML = Template_topic_c_attach_item({
						ID: data.elems[hash+"."+ext],
						ImgSrc: isImage ? e.target.result : "",
						Path: hash+"."+ext,
						FullPath: "//" + window.location.host + "/attachs/" + hash + "." + ext,
					});
					fileDock.insertBefore(fileItem,fileDock.querySelector(".attach_item_buttons"));
					
					post.classList.add("has_attachs");
					bindAttachItems();
				});
				req.open("POST","//"+window.location.host+"/"+fileDock.getAttribute("type")+"/attach/add/submit/"+fileDock.getAttribute("id"));
				req.send(formData);
			});
		} catch(e) {
			// TODO: Use a notice instead
			console.log("e:",e);
			alert(e);
		}
	}

	// Quick Topic / Quick Reply
	function uploadAttachHandler() {
		try {
			uploadFileHandler(this.files, 5, (e,filename) => {
				// TODO: Use client templates here
				let fileDock = document.getElementById("upload_file_dock");
				let fileItem = document.createElement("label");
				console.log("fileItem",fileItem);

				let ext = getExt(filename);
				// TODO: Push ImageFileExts to the client from the server in some sort of gen.js?
				let isImage = imageExts.includes(ext);
				fileItem.innerText = "." + ext;
				fileItem.className = "formbutton uploadItem";
				// TODO: Check if this is actually an image
				if(isImage) fileItem.style.backgroundImage = "url("+e.target.result+")";

				fileDock.appendChild(fileItem);
			},(e,hash, filename) => {
				console.log("hash",hash);
				let ext = getExt(filename)
				let content = document.getElementById("input_content")
				console.log("content.value", content.value);
				
				let attachItem;
				if(content.value == "") attachItem = "//" + window.location.host + "/attachs/" + hash + "." + ext;
				else attachItem = "\r\n//" + window.location.host + "/attachs/" + hash + "." + ext;
				content.value = content.value + attachItem;
				console.log("content.value", content.value);
				
				// For custom / third party text editors
				attachItemCallback(attachItem);
			});
		} catch(e) {
			// TODO: Use a notice instead
			console.log("e:",e);
			alert(e);
		}
	}

	let uploadFiles = document.getElementById("upload_files");
	if(uploadFiles != null) {
		uploadFiles.addEventListener("change", uploadAttachHandler, false);
	}
	let uploadFilesOp = document.getElementById("upload_files_op");
	if(uploadFilesOp != null) {
		uploadFilesOp.addEventListener("change", uploadAttachHandler2, false);
	}
		
	function bindAttachManager() {
		let uploadFiles = document.getElementsByClassName("upload_files_post");
		if(uploadFiles == null) return;
		for(let i = 0; i < uploadFiles.length; i++) {
			let uploader = uploadFiles[i];
			uploader.value = "";
			uploader.removeEventListener("change", uploadAttachHandler2, false);
			uploader.addEventListener("change", uploadAttachHandler2, false);
		}
	}
	bindAttachManager();
		
	function bindAttachItems() {
		$(".attach_item_select").unbind("click");
		$(".attach_item_copy").unbind("click");
		$(".attach_item_select").click(function(){
			let hold = $(this).closest(".attach_item");
			if(hold.hasClass("attach_item_selected")) hold.removeClass("attach_item_selected");
			else hold.addClass("attach_item_selected");
		});
		$(".attach_item_copy").click(function(){
			let hold = $(this).closest(".attach_item");
			let pathNode = hold.find(".attach_item_path");
			copyToClipboard(pathNode.attr("fullPath"));
		});
	}
	bindAttachItems();
		
	$(".attach_item_delete").click(function(){
		let formData = new URLSearchParams();
		formData.append("s",me.User.S);
	
		let post = this.closest(".post_item");
		let aidList = "";
		let elems = post.getElementsByClassName("attach_item_selected");
		if(elems == null) return;
			
		for(let i = 0; i < elems.length; i++) {
			let pathNode = elems[i].querySelector(".attach_item_path");
			console.log("pathNode",pathNode);
			aidList += pathNode.getAttribute("aid") + ",";
			elems[i].remove();
		}
		if(aidList.length > 0) aidList = aidList.slice(0, -1);
		console.log("aidList",aidList)
		formData.append("aids",aidList);
	
		let ec = 0;
		let e = post.getElementsByClassName("attach_item_item");
		if(e!=null) ec = e.length;
		if(ec==0) post.classList.remove("has_attachs");
			
		let req = new XMLHttpRequest();
		let fileDock = this.closest(".attach_edit_bay");
		req.open("POST","//"+window.location.host+"/"+fileDock.getAttribute("type")+"/attach/remove/submit/"+fileDock.getAttribute("id"),true);
		req.send(formData);
	
		bindAttachItems();
		bindAttachManager();
	});
		
	$(".moderate_link").click((event) => {
		event.preventDefault();
		$(".pre_opt").removeClass("auto_hide");
		$(".moderate_link").addClass("moderate_open");
		$(".topic_row").each(function(){
			$(this).click(function(){
				selectedTopics.push(parseInt($(this).attr("data-tid"),10));
				if(selectedTopics.length==1) {
					var msg = phraseBox["topic_list"]["topic_list.what_to_do_single"];
				} else {
					var msg = "What do you want to do with these "+selectedTopics.length+" topics?";
				}
				$(".mod_floater_head span").html(msg);
				$(this).addClass("topic_selected");
				$(".mod_floater").removeClass("auto_hide");
			});
		});
	
		let bulkActionSender = function(action, selectedTopics, fragBit) {
			let url = "/topic/"+action+"/submit/"+fragBit+"?s=" + me.User.S;
			$.ajax({
				url: url,
				type: "POST",
				data: JSON.stringify(selectedTopics),
				contentType: "application/json",
				error: ajaxError,
				success: () => {
					window.location.reload();
				}
			});
		};
		$(".mod_floater_submit").click(function(event){
			event.preventDefault();
			let selectNode = this.form.querySelector(".mod_floater_options");
			let optionNode = selectNode.options[selectNode.selectedIndex];
			let action = optionNode.getAttribute("value");
	
			// Handle these specially
			switch(action) {
				case "move":
					console.log("move action");
					let modTopicMover = $("#mod_topic_mover");
					$("#mod_topic_mover").removeClass("auto_hide");
					$("#mod_topic_mover .pane_row").click(function(){
						modTopicMover.find(".pane_row").removeClass("pane_selected");
						let fid = this.getAttribute("data-fid");
						if (fid == null) return;
						this.classList.add("pane_selected");
						console.log("fid: " + fid);
						forumToMoveTo = fid;
	
						$("#mover_submit").unbind("click");
						$("#mover_submit").click(function(event){
							event.preventDefault();
							bulkActionSender("move",selectedTopics,forumToMoveTo);
						});
					});
					return;
			}
				
			bulkActionSender(action,selectedTopics,"");
		});
	});
	
	function addPollInput() {
		console.log("clicked on pollinputinput");
		let dataPollInput = $(this).parent().attr("data-pollinput");
		console.log("dataPollInput: ", dataPollInput);
		if(dataPollInput == undefined) return;
		if(dataPollInput != (pollInputIndex-1)) return;
		$(".poll_content_row .formitem").append(Template_topic_c_poll_input({
			Index: pollInputIndex,
			Place: phraseBox["topic"]["topic.reply_add_poll_option"].replace("%d",pollInputIndex),
		}));
		pollInputIndex++;
		console.log("new pollInputIndex: ", pollInputIndex);
		$(".pollinputinput").off("click");
		$(".pollinputinput").click(addPollInput);
	}
	
	var pollInputIndex = 1;
	$("#add_poll_button").click((event) => {
		event.preventDefault();
		$(".poll_content_row").removeClass("auto_hide");
		$("#has_poll_input").val("1");
		$(".pollinputinput").click(addPollInput);
	});
	});
})();