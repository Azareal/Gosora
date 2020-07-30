// TODO: Push ImageFileExts to the client from the server in some sort of gen.js?
var imageExts = ["png","jpg","jpe","jpeg","jif","jfi","jfif","svg","bmp","gif","tiff","tif","webp"];

(() => {
	function copyToClipboard(str) {
		const el=document.createElement('textarea');
		el.value=str;
		el.setAttribute('readonly','');
		el.style.position='absolute';
		el.style.left='-9999px';
		document.body.appendChild(el);
		el.select();
		document.execCommand('copy');
		document.body.removeChild(el);
	}

	function uploadFileHandler(fileList, maxFiles=5, step1 = () => {}, step2 = () => {}) {
		let files = [];
		for(var i=0; i<fileList.length && i<5; i++) files[i] = fileList[i];
	
		let totalSize = 0;
		for(let i=0; i<files.length; i++) {
			log("file "+i,files[i]);
			totalSize += files[i]["size"];
		}
		if(totalSize > me.Site.MaxReqSize) throw("You can't upload this much at once, max: "+me.Site.MaxReqSize);
	
		for(let i=0; i<files.length; i++) {
			let fname = files[i]["name"];
			let f = e => {
				step1(e,fname)
					
				let reader = new FileReader();
				reader.onload = e2 => {
					crypto.subtle.digest('SHA-256',e2.target.result)
						.then(hash => {
							const hashArray = Array.from(new Uint8Array(hash))
							return hashArray.map(b => ('00' + b.toString(16)).slice(-2)).join('')
						}).then(hash => step2(e,hash,fname));
				}
				reader.readAsArrayBuffer(files[i]);
			};
				
			let ext = getExt(fname);
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
			(e,hash,fname) => {
				log("hash",hash);
				let formData = new FormData();
				formData.append("s",me.User.S);
				for(let i=0; i<this.files.length; i++) formData.append("upload_files",this.files[i]);
				bindAttachManager();

				let req = new XMLHttpRequest();
				req.addEventListener("load", () => {
					let data = JSON.parse(req.responseText);
					//log("rdata",data);
					let fileItem = document.createElement("div");
					let ext = getExt(fname);
					// TODO: Push ImageFileExts to the client from the server in some sort of gen.js?
					let isImage = imageExts.includes(ext);
					let c = "";
					if(isImage) c = " attach_image_holder"
					fileItem.className = "attach_item attach_item_item"+c;
					fileItem.innerHTML = Tmpl_topic_c_attach_item({
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
			log("e",e);
			alert(e);
		}
	}

	// Quick Topic / Quick Reply
	function uploadAttachHandler() {
		try {
			uploadFileHandler(this.files,5,(e,fname) => {
				// TODO: Use client templates here
				let fileDock = document.getElementById("upload_file_dock");
				let fileItem = document.createElement("label");
				log("fileItem",fileItem);

				let ext = getExt(fname);
				// TODO: Push ImageFileExts to the client from the server in some sort of gen.js?
				let isImage = imageExts.includes(ext);
				fileItem.innerText = "."+ext;
				fileItem.className = "formbutton uploadItem";
				// TODO: Check if this is actually an image
				if(isImage) fileItem.style.backgroundImage = "url("+e.target.result+")";

				fileDock.appendChild(fileItem);
			},(e,hash,fname) => {
				log("hash",hash);
				let ext = getExt(fname)
				let con = document.getElementById("input_content")
				log("con.value",con.value);
				
				let attachItem;
				if(con.value=="") attachItem = "//"+window.location.host+"/attachs/"+hash+"."+ext;
				else attachItem = "\r\n//"+window.location.host+"/attachs/"+hash+"."+ext;
				con.value = con.value + attachItem;
				log("con.value",con.value);
				
				// For custom / third party text editors
				attachItemCallback(attachItem);
			});
		} catch(e) {
			// TODO: Use a notice instead
			log("e",e);
			alert(e);
		}
	}

	function bindAttachManager() {
		let uploadFiles = document.getElementsByClassName("upload_files_post");
		if(uploadFiles==null) return;
		for(let i=0; i<uploadFiles.length; i++) {
			let uploader = uploadFiles[i];
			uploader.value = "";
			uploader.removeEventListener("change", uploadAttachHandler2, false);
			uploader.addEventListener("change", uploadAttachHandler2, false);
		}
	}
	
	//addInitHook("before_init_bind_page", () => {
	//log("in member.js before_init_bind_page")
	addInitHook("end_bind_topic", () => {
	log("in member.js end_bind_topic")

	let changeListener = (files,handler) => {
		if(files!=null) {
			files.removeEventListener("change", handler, false);
			files.addEventListener("change", handler, false);
		}
	};
	let uploadFiles = document.getElementById("upload_files");
	changeListener(uploadFiles,uploadAttachHandler);
	let uploadFilesOp = document.getElementById("upload_files_op");
	changeListener(uploadFilesOp,uploadAttachHandler2);
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
	
	$(".attach_item_delete").unbind("click");
	$(".attach_item_delete").click(function(){
		let formData = new URLSearchParams();
		formData.append("s",me.User.S);
	
		let post = this.closest(".post_item");
		let aidList = "";
		let elems = post.getElementsByClassName("attach_item_selected");
		if(elems==null) return;
			
		for(let i = 0; i < elems.length; i++) {
			let pathNode = elems[i].querySelector(".attach_item_path");
			log("pathNode",pathNode);
			aidList += pathNode.getAttribute("aid")+",";
			elems[i].remove();
		}
		if(aidList.length > 0) aidList = aidList.slice(0, -1);
		log("aidList",aidList)
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
	
	function addPollInput() {
		log("clicked on pollinputinput");
		let dataPollInput = $(this).parent().attr("data-pollinput");
		log("dataPollInput",dataPollInput);
		if(dataPollInput==undefined) return;
		if(dataPollInput!=(pollInputIndex-1)) return;
		$(".poll_content_row .formitem").append(Tmpl_topic_c_poll_input({
			Index: pollInputIndex,
			Place: phraseBox["topic"]["topic.reply_add_poll_option"].replace("%d",pollInputIndex),
		}));
		pollInputIndex++;
		log("new pollInputIndex",pollInputIndex);
		$(".pollinputinput").off("click");
		$(".pollinputinput").click(addPollInput);
	}
	
	let pollInputIndex = 1;
	$("#add_poll_button").unbind("click");
	$("#add_poll_button").click(ev => {
		ev.preventDefault();
		$(".poll_content_row").removeClass("auto_hide");
		$("#has_poll_input").val("1");
		$(".pollinputinput").click(addPollInput);
	});
	});
	//});
	function modCancel() {
		log("enter modCancel");
		if(!$(".mod_floater").hasClass("auto_hide")) $(".mod_floater").addClass("auto_hide")
		$(".moderate_link").unbind("click");
		$(".moderate_link").removeClass("moderate_open");
		$(".pre_opt").addClass("auto_hide");
		$(".mod_floater_submit").unbind("click");
		$("#topicsItemList,#forumItemList").removeClass("topics_moderate");
		$(".topic_selected").removeClass("topic_selected");
		// ! Be careful not to trample bindings elsewhere
		$(".topic_row").unbind("click");
		$("#mod_topic_mover").addClass("auto_hide");
	}
	function modCancelBind() {
		log("enter modCancelBind")
		$(".moderate_link").unbind("click");
		$(".moderate_open").click(ev => {
			modCancel();
			$(".moderate_open").unbind("click");
			modLinkBind();
		});
	}
	function modLinkBind() {
		log("enter modLinkBind");
		$(".moderate_link").click(ev => {
			log("enter .moderate_link");
			ev.preventDefault();
			$(".pre_opt").removeClass("auto_hide");
			$(".moderate_link").addClass("moderate_open");
			selectedTopics=[];
			modCancelBind();
			$("#topicsItemList,#forumItemList").addClass("topics_moderate");
			$(".topic_row").each(function(){
				$(this).click(function(){
					if(!this.classList.contains("can_mod")) return;
					let tid = parseInt($(this).attr("data-tid"),10);
					let sel = this.classList.contains("topic_selected");
					if(sel) {
						for(var i=0; i<selectedTopics.length; i++){
							if(selectedTopics[i]===tid) selectedTopics.splice(i, 1);
						}
					} else selectedTopics.push(tid);
					if(selectedTopics.length==1) {
						var msg = phraseBox["topic_list"]["topic_list.what_to_do_single"];
					} else {
						var msg = "What do you want to do with these "+selectedTopics.length+" topics?";
					}
					$(".mod_floater_head span").html(msg);
					if(!sel) {
						$(this).addClass("topic_selected");
						$(".mod_floater").removeClass("auto_hide");
					} else {
						$(this).removeClass("topic_selected");
					}
					if(selectedTopics.length==0 && !$(".mod_floater").hasClass("auto_hide")) $(".mod_floater").addClass("auto_hide");
				});
			});
			
			let bulkActionSender = (action,selectedTopics,fragBit) => {
				$.ajax({
					url: "/topic/"+action+"/submit/"+fragBit+"?s="+me.User.S,
					type: "POST",
					data: JSON.stringify(selectedTopics),
					contentType: "application/json",
					error: ajaxError,
					success: () => {
						window.location.reload();
					}
				});
			};
			// TODO: Should we unbind this here to avoid binding multiple listeners to this accidentally?
			$(".mod_floater_submit").click(function(ev){
				ev.preventDefault();
				let selectNode = this.form.querySelector(".mod_floater_options");
				let optionNode = selectNode.options[selectNode.selectedIndex];
				let action = optionNode.getAttribute("value");
				
				// Handle these specially
				switch(action) {
					case "move":
						log("move action");
						let modTopicMover = $("#mod_topic_mover");
						$("#mod_topic_mover").removeClass("auto_hide");
						$("#mod_topic_mover .pane_row").unbind("click");
						$("#mod_topic_mover .pane_row").click(function(){
							modTopicMover.find(".pane_row").removeClass("pane_selected");
							let fid = this.getAttribute("data-fid");
							if(fid==null) return;
							this.classList.add("pane_selected");
							log("fid",fid);
							forumToMoveTo = fid;
							
							$("#mover_submit").unbind("click");
							$("#mover_submit").click(ev => {
								ev.preventDefault();
								bulkActionSender("move",selectedTopics,forumToMoveTo);
							});
						});
						return;
				}

				bulkActionSender(action,selectedTopics,"");
			});
		});
	}
	//addInitHook("after_init_bind_page", () => {
	//addInitHook("before_init_bind_page", () => {
	//log("in member.js before_init_bind_page 2")
	addInitHook("end_bind_page", () => {
		log("in member.js end_bind_page")
		modCancel();
		modLinkBind();
	});
	addInitHook("after_init_bind_page", () => addHook("end_unbind_page", () => modCancel()))
	//});
})()