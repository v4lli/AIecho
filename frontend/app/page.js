'use client';
import { Settings } from '@/app/components/Settings';
import { useEffect, useState } from 'react';
import { VideoCapture } from '@/app/components/VideoCapture';
import { Controls } from '@/app/components/Controls';
import { AudioPlayer } from '@/app/components/AudioPlayer';
import { Timeline } from '@/app/components/Timeline';
import { useDispatch } from "react-redux";
import { ADD_MESSAGE_DESC, ADD_MESSAGE_URGENT_DESC, RESET_DESC } from "@/actions/DescriptionActions";
import { ADD_MESSAGE_TL, RESET_TL } from "@/actions/TimelineActions";
import { Desktop, MobileLandscape, MobilePortrait } from "@/app/components/Responsive";
import BoltIcon from "@mui/icons-material/Bolt";
import {Info, InfoOutlined, InfoSharp, InfoTwoTone} from "@mui/icons-material";

export default function Home() {
    const dispatch = useDispatch();
    const [videoDeviceID, setVideoDeviceID] = useState(null);
    const [isStreaming, setIsStreaming] = useState(false);
    const [upPeer, setUpPeer] = useState(null);
    const [fastMode, setFastMode] = useState(false);
    const [isClient, setIsClient] = useState(false);

    useEffect(() => {
        setUpPeer(new RTCPeerConnection());
    }, []);

    useEffect(() => {
        setIsClient(true)
    },[]);

    function sortByMimeTypes(codecs, preferredOrder) {
        return codecs.sort((a, b) => {
            const indexA = preferredOrder.indexOf(a.mimeType);
            const indexB = preferredOrder.indexOf(b.mimeType);
            const orderA = indexA >= 0 ? indexA : Number.MAX_VALUE;
            const orderB = indexB >= 0 ? indexB : Number.MAX_VALUE;
            return orderA - orderB;
        });
    }

    const handleStart = () => {
        if (videoDeviceID && !isStreaming) {
            navigator.mediaDevices.getUserMedia({
                audio: false, video: { deviceId: videoDeviceID }
            }).then(stream => {
                stream.getTracks().forEach(track => {
                    upPeer.addTrack(track, stream);
                });
            });


            upPeer.onnegotiationneeded = async () => {
                try {
                    const [transceiver1] = upPeer.getTransceivers();
                    const codecs1 = RTCRtpReceiver.getCapabilities("video").codecs;
                    transceiver1.setCodecPreferences(sortByMimeTypes(codecs1, ["video/VP8"])); // <---
                    console.log(`pc1 prefers ${[...new Set(codecs1.map(({ mimeType }) => mimeType))]}`);
                    let sendChannel = upPeer.createDataChannel('foo')
                    sendChannel.onmessage = e => {
                        let message = JSON.parse(e.data)
                        switch (message.type) {
                            case "desc":
                                if (message.urgent) {
                                    dispatch({ type: ADD_MESSAGE_URGENT_DESC, payload: message.content })
                                } else {
                                    dispatch({ type: ADD_MESSAGE_DESC, payload: message.content })
                                }
                                break
                            case "tl":
                                dispatch({ type: ADD_MESSAGE_TL, payload: message.content })
                                if (fastMode) {
                                    dispatch({ type: ADD_MESSAGE_DESC, payload: message.content })
                                }
                        }
                    }
                    const offer = await upPeer.createOffer();
                    await upPeer.setLocalDescription(offer);
                    upPeer.onicegatheringstatechange = async () => {
                        if (upPeer.iceGatheringState === 'complete') {
                            let localSDP = upPeer.localDescription.sdp;
                            localSDP += "\r\na=fastMode:${fastMode? 1:0}\r\n"
                            try {
                                let UUID = self.crypto.randomUUID();
                                const response = await fetch('https://api.aiecho.unimplemented.org/wish/whip/' + UUID + `/${fastMode ? 'fast' : 'normal'}`, {
                                    method: 'POST', headers: {
                                        'Content-Type': 'application/sdp'
                                    }, body: localSDP
                                });
                                const remoteSDP = await response.text();
                                const remoteDesc = new RTCSessionDescription({
                                    type: 'answer', sdp: remoteSDP
                                });
                                await upPeer.setRemoteDescription(remoteDesc);
                            } catch (error) {
                                console.error('Error sending offer to the server:', error);
                            }
                        }
                    };
                } catch (error) {
                    console.error('Failed to create offer:', error);
                }
            };
            setIsStreaming(true);
        }
    };
    const handleStop = () => {
        if (upPeer) {
            upPeer.close();
            setUpPeer(new RTCPeerConnection());
            setIsStreaming(false);
            dispatch({ type: RESET_TL, payload: null });
            dispatch({ type: RESET_DESC, payload: null })
        }
    };

    const toggleFastMode = () => {
        handleStop();
        setFastMode(!fastMode);
    }

    return (<div className="flex flex-box">
        {isClient && (<>
            <Desktop>
                <div className="main-content">
                    <Timeline className="timeline"/>
                    <VideoCapture className="video-capture" videoDevice={videoDeviceID} onStream={isStreaming}/>
                </div>
                <div className="flex-row controls">
                    <div className="logo-container">
                        <img src="/logo.png" alt="AIEcho Logo"/>
                    </div>
                    <Controls onStart={handleStart} onStop={handleStop}/>
                    <Settings onVideoDeviceChange={setVideoDeviceID}/>
                    <AudioPlayer toggleFastMode={toggleFastMode} onStream={isStreaming}/>
                </div>
            </Desktop>
            <MobileLandscape>
                <div className="main-content">
                    <Timeline className="timeline"/>
                    <VideoCapture className="video-capture" videoDevice={videoDeviceID} onStream={isStreaming}/>
                </div>
                <div className="controls">
                    <Controls onStart={handleStart} onStop={handleStop}/>
                    <Settings onVideoDeviceChange={setVideoDeviceID}/>
                    <AudioPlayer toggleFastMode={toggleFastMode} onStream={isStreaming}/>
                </div>
            </MobileLandscape>
            <MobilePortrait>
                <div className="main-content">
                    <Timeline className="timeline"/>
                    <VideoCapture className="video-capture" videoDevice={videoDeviceID} onStream={isStreaming}/>
                </div>
                <div className="controls">
                    <Controls onStart={handleStart} onStop={handleStop}/>
                    <Settings onVideoDeviceChange={setVideoDeviceID}/>
                    <AudioPlayer toggleFastMode={toggleFastMode} onStream={isStreaming}/>
                </div>
            </MobilePortrait></>)}
        <div className="links"><InfoOutlined className="info-icon" /> AIEcho is a technology demonstration. Don't rely on it, the AI makes stuff up. Report an
            issue or contribute on <a href="https://github.com/v4lli/AIecho">GitHub</a>. <a
                href="https://meteocool.com/imprint">Imprint</a></div>
    </div>);
}
