const localVideo = document.getElementById("local-video");
const remotesDiv = document.getElementById("remotes");

/* eslint-env browser */
const joinBtn = document.getElementById("join-btn");
const getonlineSourceBtn = document.getElementById("getonlineSource-btn");

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

const onlineSourceList = document.getElementById("onLineSourceList");
let localDataChannel;
let trackEvent;

//const url = 'http://192.168.1.199:5551';
const url = 'http://120.78.200.246:5551';
const connectType={
  Control : 0,
  View:1,
  Manage : 2,
}
const sourceType={
Car :1,
Feed :2,
Camera : 3,
Boat : 4,
Submarine : 5,//潜艇
}
//const url = 'http://www.bxzryd.cn:5551';
//const uid = uuidv4();
//const myID = "uid198";//uuidv4();
const myID=uidTime();
//const sid = "ion";
let room;
let rtc;
let localStream;
let start;
let channelDataEvent;
const getOnlineSources = async () => {
  const connector = new Ion.Connector(url, "token");
  connector.onopen = function (service){
      console.log("[onopen]: service = ", service.name);
  };

  connector.onclose = function (service){
      console.log('[onclose]: service = ' + service.name);
  };

  //getonlineSourceBtn.disabled = "true";
  rtc = new Ion.RTC(connector);
  rtc.getOnlineSources(sourceType.Car).then(function(array){
    console.log(array);
    //alert(array[0].getSid()+" "+array[0].getUid()+" "+array[1].getSid()+" "+array[1].getUid());    
    array.forEach(function(item,index){
      //alert(item.getSid()+" "+item.getUid())
      onlineSourceList.innerHTML +="<li id='"+ item.getUid()+"')'>"+item.getName()+" "+item.getUid()+"</li>";
    })
    //onlineSourceList.innerHTML="
  });
  onlineSourceList.addEventListener("click",function(event){
    var soureId = event.target.id
    //alert(sid);
    wantConnect(myID,soureId);
  })
  onlineSourceList.addEventListener("mouseover",function(event){
    event.target.style.color="red";
    //alert(name);
  })
  onlineSourceList.addEventListener("mouseout",function(event){
    event.target.style.color="black";
    //alert(name);
  })
}

function clickList(source){
  alert(source);
}

// const join = async (sid) => {
//     console.log("[join]: sid="+sid+" uid=", uid)
//     const connector = new Ion.Connector(url, "token");    
//     connector.onopen = function (service){
//         console.log("[onopen]: service = ", service.name);
//     };

//     connector.onclose = function (service){
//         console.log('[onclose]: service = ' + service.name);
//     };

//     joinBtn.disabled = "true";
//     leaveBtn.removeAttribute('disabled');
//     publishBtn.removeAttribute('disabled');
//     publishSBtn.removeAttribute('disabled');

//     rtc = new Ion.RTC(connector);

//     rtc.ontrack = (track, stream) => {
//       console.log("got ", track.kind, " track", track.id, "for stream", stream.id);
//       if (track.kind === "video") {
//         track.onunmute = () => {
//           if (!streams[stream.id]) {
//             const remoteVideo = document.createElement("video");
//             remoteVideo.srcObject = stream;
//             remoteVideo.autoplay = true;
//             remoteVideo.muted = true;
//             remoteVideo.addEventListener("loadedmetadata", function () {
//               sizeTag.innerHTML = `${remoteVideo.videoWidth}x${remoteVideo.videoHeight}`;
//             });
    
//             remoteVideo.onresize = function () {
//               sizeTag.innerHTML = `${remoteVideo.videoWidth}x${remoteVideo.videoHeight}`;
//             };
//             remotesDiv.appendChild(remoteVideo);
			
//             streams[stream.id] = stream;
//             stream.onremovetrack = () => {
//               if (streams[stream.id]) {
//                 remotesDiv.removeChild(remoteVideo);
//                 streams[stream.id] = null;
//               }
//             };
//             getStats();
//           }
//         };
//       }
//     };

//     rtc.ontrackevent = function (ev) {
//       console.log("ontrackevent: \nuid = ", ev.uid, " \nstate = ", ev.state, ", \ntracks = ", JSON.stringify(ev.tracks));
//       if (trackEvent === undefined) {
//         console.log("store trackEvent=", ev)
//         trackEvent = ev;
//       }
//       remoteSignal.innerHTML = remoteSignal.innerHTML + JSON.stringify(ev) + '\n';
//     };

//     rtc.join(sid, uid);

//     const streams = {};

//     start = (sc) => {
//       // publishSBtn.disabled = "true";
//       // publishBtn.disabled = "true";
//       //  localDataChannel = rtc.createDataChannel(uid);
//       //  localDataChannel.onopen=()=>{
//       //    localDataChannel.send("wuyaxiong235467")
//       //  }
//       //  localDataChannel.onmessage=(msg)=>{
//       //   console.log("localDataChannel.onmessage:",msg)
//       //  }

//       let constraints = {
//         resolution: resolutionBox.options[resolutionBox.selectedIndex].value,
//         codec: codecBox.options[codecBox.selectedIndex].value,
//         audio: true,
//         simulcast: sc,
//       }
//       console.log("getUserMedia constraints=", constraints)
//       Ion.LocalStream.getUserMedia(constraints)
//         .then((media) => {
//           localStream = media;
//           localVideo.srcObject = media;
//           localVideo.autoplay = true;
//           localVideo.controls = true;
//           localVideo.muted = true;

//           rtc.publish(media);
//           localDataChannel = rtc.createDataChannel(uid);
//         })
//         .catch(console.error);
     
//     };

//     rtc.ondatachannel = ({ channel }) => {
//       console.log("ondatachannel,",channel)
//       channel.send("wuyaxiong call back")
//       channel.onmessage = ({ data }) => {
//         console.log("datachannel msg:", data)

//         remoteData.innerHTML = data;
//       };
//     };
// }


const wantConnect = async (myid,detination) => {
  console.log("[join]: myid="+myid+" detination=", detination)
  const connector = new Ion.Connector(url, "token");    
  connector.onopen = function (service){
      console.log("[onopen]: service = ", service.name);
  };

  connector.onclose = function (service){
      console.log('[onclose]: service = ' + service.name);
  };

  joinBtn.disabled = "true";
  leaveBtn.removeAttribute('disabled');
  publishBtn.removeAttribute('disabled');
  publishSBtn.removeAttribute('disabled');

  rtc = new Ion.RTC(connector);

  rtc.ontrack = (track, stream) => {
    console.log("got ", track.kind, " track", track.id, "for stream", stream.id);
    if (track.kind === "video") {
      track.onunmute = () => {
        if (!streams[stream.id]) {
          const remoteVideo = document.createElement("video");
          remoteVideo.srcObject = stream;
          remoteVideo.autoplay = true;
          remoteVideo.muted = true;
          remoteVideo.addEventListener("loadedmetadata", function () {
            sizeTag.innerHTML = `${remoteVideo.videoWidth}x${remoteVideo.videoHeight}`;
          });
  
          remoteVideo.onresize = function () {
            sizeTag.innerHTML = `${remoteVideo.videoWidth}x${remoteVideo.videoHeight}`;
          };
          remotesDiv.appendChild(remoteVideo);
    
          streams[stream.id] = stream;
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

  rtc.wantConnect(myid,detination, connectType.View);

//   localDataChannel = rtc.createDataChannel(uid);
//   localDataChannel.onopen=()=> {
//             console.log("data channel onpen")
//   }
//  localDataChannel.onmessage=({msg})=> {
//   console.log("data channel msg: ", msg)
//  }

  const streams = {};

  start = (sc) => {
    // publishSBtn.disabled = "true";
    // publishBtn.disabled = "true";
    //  localDataChannel = rtc.createDataChannel(uid);
    //  localDataChannel.onopen=()=>{
    //    localDataChannel.send("wuyaxiong235467")
    //  }
    //  localDataChannel.onmessage=(msg)=>{
    //   console.log("localDataChannel.onmessage:",msg)
    //  }

    let constraints = {
      resolution: resolutionBox.options[resolutionBox.selectedIndex].value,
      codec: codecBox.options[codecBox.selectedIndex].value,
      audio: true,
      simulcast: sc,
    }
    console.log("getUserMedia constraints=", constraints)
    Ion.LocalStream.getUserMedia(constraints)
      .then((media) => {
        localStream = media;
        localVideo.srcObject = media;
        localVideo.autoplay = true;
        localVideo.controls = true;
        localVideo.muted = true;
        rtc.publish(media);
        localDataChannel = rtc.createDataChannel(uid);
      }).catch(console.error);   
  };

  rtc.ondatachannel = (ev) => {
    console.log("ondatachannel,",ev)
    channelDataEvent=ev;
    ev.channel.onmessage = ({ data }) => {
      console.log("datachannel msg:", data)
      //ev.channel.send("wuyaxiong nv call back");
      remoteData.innerHTML = data;
    };
  };
}

const send = () => {
  // if (!localDataChannel || !localDataChannel.readyState) {
  //   alert('publish first!', '', {
  //     confirmButtonText: 'OK',
  //   });
  //   return
  // }

  if (channelDataEvent.channel.readyState === "open") {
    console.log("datachannel send:", localData.value)
    channelDataEvent.channel.send(localData.value);
  } 
};

const leave = () => {
    console.log("[leave]: sid=" + sid + " uid=", uid)
    rtc.leave(sid, uid);
    joinBtn.removeAttribute('disabled');
    leaveBtn.disabled = "true";
    publishBtn.disabled = "true";
    publishSBtn.disabled = "true";
    location.reload();
}

const subscribe = () => {
    let layer = subscribeBox.value
    console.log("subscribe trackEvent=", trackEvent, "layer=", layer)
    var infos = [];
    trackEvent.tracks.forEach(t => {
        if (t.layer === layer && t.kind === "video"){
          infos.push({
            track_id: t.id,
            mute: t.muted,
            layer: t.layer,
            subscribe: true
          });
        }
        
        if (t.kind === "audio"){
          infos.push({
            track_id: t.id,
            mute: t.muted,
            layer: t.layer,
            subscribe: true
          });
        }
    });
    console.log("subscribe infos=", infos)
    rtc.subscribe(infos);
}

const controlLocalVideo = (radio) => {
  if (radio.value === "false") {
    localStream.mute("video");
  } else {
    localStream.unmute("video");
  }
};

const controlLocalAudio = (radio) => {
  if (radio.value === "false") {
    localStream.mute("audio");
  } else {
    localStream.unmute("audio");
  }
};

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

function uidTime(){
  var uid ="",random;
  const time=(new Date().getTime()-new Date(2022,0).getTime()).toString();
  for(i=0;i<8;i++) {
       random = Math.floor(Math.random()*16);
       uid+=random.toString(16);
  }
  return "web"+time+uid;
}

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