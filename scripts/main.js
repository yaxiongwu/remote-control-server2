const localVideo = document.getElementById("local-video");
const remotesDiv = document.getElementById("remotes");
const remoteVideo = document.getElementById("remote_video");

/* eslint-env browser */
const joinBtn = document.getElementById("join-btn");
const leaveBtn = document.getElementById("leave-btn");
const publishBtn = document.getElementById("publish-btn");
const publishSBtn = document.getElementById("publish-simulcast-btn");

const codecBox = document.getElementById("select-box1");
const resolutionBox = document.getElementById("select-box2");
const simulcastBox = document.getElementById("check-box");
const localData = document.getElementById("local-data");
const remoteData = document.getElementById("remote-data");
const remoteSignal= document.getElementById("remote-signal");
const subscribeBox = document.getElementById("select-box3");
const sizeTag = document.getElementById("size-tag");
const brTag = document.getElementById("br-tag");
let localDataChannel;
let trackEvent;

//const url = 'http://192.168.1.199:5551';
const url = 'http://120.78.200.246:5551';
//const url = 'http://www.bxzryd.cn:5551';
//const uid = uuidv4();
const uid = "uid198";//uuidv4();
const sid = "ion";
let room;
let rtc;
let localStream;
let start;

// window.onload=function(){
// 	//remotesDiv.style="transform: rotate(90deg);-o-transform: rotate(90deg);-webkit-transform: rotate(90deg);-moz-transform: rotate(90deg);"
//  //    remotesDiv.style.height=document.body.clientWidth+'px';
// 	// remotesDiv.style.width='800px';
// 	// remotesDiv.style.backgroundColor="#f00";
	
// 	//remotesDiv.style.transform="rotate(90deg)";
// 	//document.body.style.transform ="rotate(90deg)";
// 	console.log("document.body.clientWidth:"+document.body.clientWidth);
// 	console.log("document.body.clientHeight:"+document.body.clientHeight);
// 	console.log("document.documentElement.clientWidth:"+document.documentElement.clientWidth);
// 	console.log("document.documentElement.clientHeight:"+document.documentElement.clientHeight);
	
	
	
// }

const join = async () => { 
	joinBtn.disabled = "true";
    console.log("[join]: sid="+sid+" uid=", uid)
    const connector = new Ion.Connector(url, "token");    
    connector.onopen = function (service){
        console.log("[onopen]: service = ", service.name);
    };

    connector.onclose = function (service){
        console.log('[onclose]: service = ' + service.name);
    };

   
   
    rtc = new Ion.RTC(connector);

    rtc.ontrack = (track, stream) => {
      console.log("got ", track.kind, " track", track.id, "for stream", stream.id);
      if (track.kind === "video") {
        track.onunmute = () => {
          if (!streams[stream.id]) {
           // const remoteVideo = document.createElement("video");
            remoteVideo.srcObject = stream;
            remoteVideo.autoplay = true;
            remoteVideo.muted = true;			
			//remoteVideo.style.objectFit="cover";
            remoteVideo.addEventListener("loadedmetadata", function () {
              sizeTag.innerHTML = `${remoteVideo.videoWidth}x${remoteVideo.videoHeight}`;
            });
			
            remoteVideo.onresize = function () {
              sizeTag.innerHTML = `${remoteVideo.videoWidth}x${remoteVideo.videoHeight}`;
            };
			console.log("document.body.clientWidth:"+document.body.clientWidth);
			console.log("document.body.clientHeight:"+document.body.clientHeight);
			console.log("remotesDiv.offsetHeight:"+remotesDiv.offsetHeight);
			console.log("remotesDiv.offsetWidth:"+remotesDiv.offsetWidth);
			
			// remotesDiv.style="transform: rotate(90deg);-o-transform: rotate(90deg);-webkit-transform: rotate(90deg);-moz-transform: rotate(90deg);"
			// remotesDiv.style.height=document.body.clientWidth+'px';
			// remotesDiv.style.position="absolute";
			// remotesDiv.style.left="-100px";
			// remotesDiv.style.top="100px";
			
			console.log("document.body.clientWidth:"+document.body.clientWidth);
			console.log("document.body.clientHeight:"+document.body.clientHeight);
		    console.log("remotesDiv.offsetHeight:"+remotesDiv.offsetHeight);
		    console.log("remotesDiv.offsetWidth:"+remotesDiv.offsetWidth);
			
            remotesDiv.appendChild(remoteVideo);
            streams[stream.id] = stream;
			console.log("remoteVideo.offsetWidth:"+remoteVideo.offsetWidth);
			console.log("remoteVideo.offsetHeight:"+remoteVideo.offsetHeight);
            stream.onremovetrack = () => {
              if (streams[stream.id]) {
                remotesDiv.removeChild(remoteVideo);
                streams[stream.id] = null;
              }
            };
            getStats();
          }
        };
      }
    };

    rtc.ontrackevent = function (ev) {
      console.log("ontrackevent: \nuid = ", ev.uid, " \nstate = ", ev.state, ", \ntracks = ", JSON.stringify(ev.tracks));
      if (trackEvent === undefined) {
        console.log("store trackEvent=", ev)
        trackEvent = ev;
      }
      remoteSignal.innerHTML = remoteSignal.innerHTML + JSON.stringify(ev) + '\n';
    };

    rtc.join(sid, uid);

    const streams = {};

    start = (sc) => {
      // publishSBtn.disabled = "true";
      // publishBtn.disabled = "true";
       localDataChannel = rtc.createDataChannel(uid);
       localDataChannel.onopen=()=>{
         localDataChannel.send("wuyaxiong235467")
       }
       localDataChannel.onmessage=(msg)=>{
        console.log("localDataChannel.onmessage:",msg)
       }

      // let constraints = {
      //   resolution: resolutionBox.options[resolutionBox.selectedIndex].value,
      //   codec: codecBox.options[codecBox.selectedIndex].value,
      //   audio: true,
      //   simulcast: sc,
      // }
      // console.log("getUserMedia constraints=", constraints)
      // Ion.LocalStream.getUserMedia(constraints)
      //   .then((media) => {
      //     localStream = media;
      //     localVideo.srcObject = media;
      //     localVideo.autoplay = true;
      //     localVideo.controls = true;
      //     localVideo.muted = true;

      //     rtc.publish(media);
      //     localDataChannel = rtc.createDataChannel(uid);
      //   })
      //   .catch(console.error);
     
    };

    rtc.ondatachannel = ({ channel }) => {
      console.log("ondatachannel,",channel)
      channel.onmessage = ({ data }) => {
        console.log("datachannel msg:", data)
        remoteData.innerHTML = data;
      };
    };
}

const send = () => {
  if (!localDataChannel || !localDataChannel.readyState) {
    alert('publish first!', '', {
      confirmButtonText: 'OK',
    });
    return
  }

  if (localDataChannel.readyState === "open") {
    console.log("datachannel send:", localData.value)
    localDataChannel.send(localData.value);
  }
};

const leave = () => {
    console.log("[leave]: sid=" + sid + " uid=", uid)
    rtc.leave(sid, uid);
    joinBtn.removeAttribute('disabled');
    // leaveBtn.disabled = "true";
    // publishBtn.disabled = "true";
    // publishSBtn.disabled = "true";
    // location.reload();
}

const subscribe = () => {
    // let layer = subscribeBox.value
    // console.log("subscribe trackEvent=", trackEvent, "layer=", layer)
    // var infos = [];
    // trackEvent.tracks.forEach(t => {
    //     if (t.layer === layer && t.kind === "video"){
    //       infos.push({
    //         track_id: t.id,
    //         mute: t.muted,
    //         layer: t.layer,
    //         subscribe: true
    //       });
    //     }
        
    //     if (t.kind === "audio"){
    //       infos.push({
    //         track_id: t.id,
    //         mute: t.muted,
    //         layer: t.layer,
    //         subscribe: true
    //       });
    //     }
    // });
    // console.log("subscribe infos=", infos)
    // rtc.subscribe(infos);
}



// const controlLocalVideo = (radio) => {
//   if (radio.value === "false") {
//     localStream.mute("video");
//   } else {
//     localStream.unmute("video");
//   }
// };

// const controlLocalAudio = (radio) => {
//   if (radio.value === "false") {
//     localStream.mute("audio");
//   } else {
//     localStream.unmute("audio");
//   }
// };

const getStats = () => {
  let bytesPrev;
  let timestampPrev;
  setInterval(() => {
    rtc.getSubStats(null).then((results) => {
      results.forEach((report) => {
        const now = report.timestamp;

        let bitrate;
        if (
          report.type === "inbound-rtp" &&
          report.mediaType === "video"
        ) {
          const bytes = report.bytesReceived;
          if (timestampPrev) {
            bitrate = (8 * (bytes - bytesPrev)) / (now - timestampPrev);
            bitrate = Math.floor(bitrate);
          }
          bytesPrev = bytes;
          timestampPrev = now;
        }
        if (bitrate) {
          brTag.innerHTML = `${bitrate} kbps @ ${report.framesPerSecond} fps`;
        }
      });
    });
  }, 1000);
};

function syntaxHighlight(json) {
  json = json
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;");
  return json.replace(
    /("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g,
    function (match) {
      let cls = "number";
      if (/^"/.test(match)) {
        if (/:$/.test(match)) {
          cls = "key";
        } else {
          cls = "string";
        }
      } else if (/true|false/.test(match)) {
        cls = "boolean";
      } else if (/null/.test(match)) {
        cls = "null";
      }
      return '<span class="' + cls + '">' + match + "</span>";
    }
  );
}