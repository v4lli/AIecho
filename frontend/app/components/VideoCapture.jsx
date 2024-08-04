import {useEffect, useRef} from 'react';

export const VideoCapture = ({onStream, videoDevice}) => {
    const videoRef = useRef(null);
    const peerRef = useRef(null);
    useEffect(() => {
        const startStream = async () => {
            try {
                videoRef.current.srcObject = (await navigator.mediaDevices.getUserMedia({
                    audio: false, video: {deviceId: videoDevice}
                }));
                videoRef.current.play();
            } catch (err) {
                console.error('Error accessing media devices:', err);
            }
        };
        if (videoDevice) {
            startStream();
        }
    }, [videoDevice, onStream]);
    return (<div className="video_container"><video aria-label={'Control Window for Video Upstream'} ref={videoRef} muted playsInline tabIndex={-1}/></div> );
};
